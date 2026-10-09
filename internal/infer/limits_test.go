package infer_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
	"custos/internal/testbudget"
)

// returnTypes infers the type of every `return` expression of src.
func returnTypes(t *testing.T, src string) []string {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
	var out []string
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if r, ok := n.(*syntax.Return); ok && r.Expr != nil {
			out = append(out, env.TypeOf(r.Expr).DocString())
		}
		return true
	})
	return out
}

// A hostile @extends chain whose arguments use the template several times
// grows the bindings geometrically; the walk and the binding size are
// capped (security review 2026-10-07: 1,000 classes took 17 s per call).
func TestGenericChainBounded(t *testing.T) {
	var b strings.Builder
	b.WriteString("<?php\n/** @template T */\nclass C0 { /** @return T */ public function m() {} }\n")
	const n = 1000
	for i := 1; i <= n; i++ {
		fmt.Fprintf(&b, "/**\n * @template T\n * @extends C%d<array{a: T, b: T, c: T, d: list<T>}>\n */\nclass C%d extends C%d {}\n", i-1, i, i-1)
	}
	// Distinct receivers defeat the per-receiver cache.
	for i := 0; i < 20; i++ {
		fmt.Fprintf(&b, "function f%d(){ /** @var C%d<array{k%d: int}> $x */ $x = g(); return $x->m(); }\n", i, n-i, i)
	}
	// A short chain still binds through several levels.
	b.WriteString("/**\n * @template T\n * @extends C0<list<T>>\n */\nclass S1 extends C0 {}\n")
	b.WriteString("/**\n * @template T\n * @extends S1<T>\n */\nclass S2 extends S1 {}\n")
	b.WriteString("function short(){ /** @var S2<int> $x */ $x = g(); return $x->m(); }\n")
	start := time.Now()
	got := returnTypes(t, b.String())
	if d := time.Since(start); d > testbudget.Of(2*time.Second) && !raceEnabled {
		t.Errorf("generic chain took %v", d)
	}
	for _, r := range got {
		if len(r) > 4*4096 {
			t.Errorf("binding grew to %d bytes", len(r))
		}
	}
	if last := got[len(got)-1]; last != "array<int,int>" {
		t.Errorf("short chain: got %s, want array<int,int>", last)
	}
}

// Doc names are resolved against the templates and aliases of the
// enclosing declarations; their doc comments are parsed once per file, not
// once per inline @var (an 800 KB class doc and 5,000 annotations took 12 s
// when re-parsed per annotation). Four times the annotations must cost about
// four times as much (linear), not sixteen (quadratic): the ratio holds on a
// loaded machine where a fixed limit does not.
func TestScopeDocsParsedOnce(t *testing.T) {
	small, large := scopeDocsTime(t, 1250), scopeDocsTime(t, 5000)
	ratio := float64(large) / float64(small)
	t.Logf("1250 annotations %v, 5000 %v (x%.1f)", small, large, ratio)
	if ratio > 9 && !raceEnabled {
		t.Errorf("4x more annotations took x%.1f longer (super-linear)", ratio)
	}
	if large > testbudget.Of(30*time.Second) {
		t.Errorf("took %v", large)
	}
}

// scopeDocsTime types n annotated variables under a large class doc and
// returns the best of two timings.
func scopeDocsTime(t *testing.T, n int) time.Duration {
	var b strings.Builder
	b.WriteString("<?php\n/**\n")
	for i := 0; i < 200; i++ {
		fmt.Fprintf(&b, " * @phpstan-type A%d array{k: %s}\n", i, strings.Repeat("int|", 1000)+"int")
	}
	b.WriteString(" */\nclass C {\n    public function m() {\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&b, "        /** @var A%d $v%d */\n        $v%d = g();\n", i%200, i, i)
	}
	fmt.Fprintf(&b, "        return $v%d;\n    }\n}\n", n-1)
	best := time.Duration(1<<63 - 1)
	for run := 0; run < 2; run++ {
		start := time.Now()
		got := returnTypes(t, b.String())
		if d := time.Since(start); d < best {
			best = d
		}
		if len(got) != 1 || got[0] != "array{k:int}" {
			t.Errorf("n=%d: got %v", n, got)
		}
	}
	return best
}

// typeAll types every variable and the T-rules type of every assigned
// value of src, returning the elapsed time.
func typeAll(t *testing.T, src string) time.Duration {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	env := infer.NewEnv(f, names.New(f), ix, phpver.PHP84)
	tr := infer.NewTRules(env)
	start := time.Now()
	syntax.InspectFile(f, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Variable, *syntax.PropertyFetch, *syntax.ArrayDimFetch:
			env.TypeOf(n.(syntax.Expr))
		case *syntax.Assign:
			tr.TypeOf(n.Value)
		}
		return true
	})
	return time.Since(start)
}

// Reads used to rescan the statements before them (guards), the
// variable's definitions and, for T-rules, the whole scope: long bodies
// assigning or guarding one variable thousands of times were quadratic
// (security review 2026-10-07: 20k `$a['k'] = …` 11 s, 20k `$o = new X`
// 9 s, 20k `$s = str_replace(…, $s)` 3.5 s). A class with 20k properties
// made untyped property inference quadratic.
func TestLongBodiesBounded(t *testing.T) {
	const n = 20000
	body := func(line string) string {
		return "<?php\nclass X { public function m(): int { return 1; } }\nfunction f(array $arr, string $s, $x) {\n $a = [];\n" +
			strings.Repeat(line+"\n", n) + " return $a;\n}\n"
	}
	cases := map[string]string{
		"element writes": body(" $a['k'] = $a['k'] . 'x';"),
		"new objects":    body(" $o = new X; $o->m();"),
		"str_replace":    body(" $s = str_replace('a', 'b', $s);"),
		"array_push":     body(" array_push($arr, $x);"),
		"guards":         body(" if ($x === null) { return; }"),
		"isset guards":   body(" if (!isset($arr['k'])) { return; } $y = $arr['k'];"),
	}
	var props strings.Builder
	props.WriteString("<?php\nfinal class P {\n")
	for i := 0; i < n; i++ {
		fmt.Fprintf(&props, " private $p%d = %d;\n", i, i)
	}
	for i := 0; i < n; i += 10 {
		fmt.Fprintf(&props, " public function s%d() { $this->p%d = 2; return $this->p%d; }\n", i, i, i)
	}
	props.WriteString("}\n")
	cases["properties"] = props.String()
	for name, src := range cases {
		if d := typeAll(t, src); d > testbudget.Of(2*time.Second) && !raceEnabled {
			t.Errorf("%s: took %v", name, d)
		}
	}
}

// Past maxVarDefs definitions a variable is unknown; below, unchanged.
func TestManyDefinitions(t *testing.T) {
	few := "<?php\nfunction f() {\n" + strings.Repeat(" $v = 1;\n", 10) + " return $v;\n}\n"
	many := "<?php\nfunction f() {\n" + strings.Repeat(" $v = 1;\n", 600) + " return $v;\n}\n"
	if got := returnTypes(t, few); len(got) != 1 || got[0] != "int" {
		t.Errorf("few: got %v", got)
	}
	if got := returnTypes(t, many); len(got) != 1 || got[0] != "?unknown" {
		t.Errorf("many: got %v", got)
	}
}
