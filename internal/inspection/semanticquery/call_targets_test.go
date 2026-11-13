package semanticquery

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/stubs"
)

// probeFrozen runs check on the `probe(...)` calls of src analyzed against
// a frozen project index holding src, as a project run does.
func probeFrozen(tb testing.TB, src string, check func(*analysis.Context, *syntax.FuncCall)) {
	tb.Helper()
	parse := func() *syntax.File { return syntax.Parse("frozen.php", []byte(src), syntax.Options{}) }
	indexed := parse()
	fs := index.Extract(indexed)
	fs.Digest = index.Digest(indexed)
	project := index.New(stubs.Index())
	project.Add(fs)
	project.Freeze()
	ran := false
	p := nativeProbe{check: func(ctx *analysis.Context, c *syntax.FuncCall) {
		if name, ok := c.Name.(*syntax.Name); ok && name.Value == "probe" {
			if ctx.Index() != project {
				tb.Fatal("the file got an index layer of its own")
			}
			ran = true
			check(ctx, c)
		}
	}}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.ID()}})
	if err != nil {
		tb.Fatal(err)
	}
	e.WithIndex(project).Analyze(parse())
	if !ran {
		tb.Fatal("no probe() call checked")
	}
}

func TestCallTargetsTrustEachCallerOnASharedIndex(t *testing.T) {
	probeFrozen(t, `<?php
class Bag extends ArrayObject {}
class Sack extends Bag {}
function probe($x) {}
function run(Bag $b) { probe($b->count()); }`, func(ctx *analysis.Context, c *syntax.FuncCall) {
		call := CallArgument(c.Args, 0, "")
		callee, ok := ResolveCallee(ctx, call)
		if !ok {
			t.Fatal("callee unresolved")
		}
		trust := func(file, name string) bool { return true }
		distrust := func(file, name string) bool { return false }
		for range 2 {
			if name, targets, ok := CallTargets(ctx, call, callee, trust); !ok || name != "ArrayObject::count" || len(targets) != 2 || !targets[0].Builtin {
				t.Fatalf("trusted: %q %+v %v", name, targets, ok)
			}
			if _, _, ok := CallTargets(ctx, call, callee, distrust); ok {
				t.Fatal("an untrusted builtin was listed")
			}
		}
	})
}

func TestCallTargetsOfSeveralReceiversNeverShareStorage(t *testing.T) {
	// Both unions start with the shared targets of A: listing the second
	// must leave the first as it was.
	var first []CallTarget
	probeFrozen(t, `<?php
class A { function f($x) {} }
class B { function f($x, $y) {} }
class C { function f($x, $y, $z) {} }
function probe($x) {}
/**
 * @param A|B $o
 * @param A|C $p
 */
function run($o, $p) { probe($o->f(1)); probe($p->f(1)); }`, func(ctx *analysis.Context, c *syntax.FuncCall) {
		call := CallArgument(c.Args, 0, "")
		callee, _ := ResolveCallee(ctx, call)
		_, targets, ok := CallTargets(ctx, call, callee, func(string, string) bool { return true })
		if !ok || len(targets) != 2 || len(targets[0].Params) != 1 {
			t.Fatalf("targets %+v", targets)
		}
		if first == nil {
			first = targets
		}
		if len(first[1].Params) != 2 {
			t.Fatalf("the targets of A|B became %+v", first)
		}
	})
}

func TestCallTargetsGiveUpOnAnUncertainOverride(t *testing.T) {
	// Kid hands the protected method to __call.
	probeFrozen(t, `<?php
class Base { protected function f($x) {} function run() { probe($this->f(1)); } }
class Kid extends Base { function __call($name, $args) {} }
function probe($x) {}`, func(ctx *analysis.Context, c *syntax.FuncCall) {
		call := CallArgument(c.Args, 0, "")
		callee, _ := ResolveCallee(ctx, call)
		if _, targets, ok := CallTargets(ctx, call, callee, func(string, string) bool { return true }); ok {
			t.Fatalf("targets %+v", targets)
		}
	})
}

// BenchmarkCallTargetsOverrides lists the targets of a method call on an
// abstract class that 60 subclasses (in three levels) override, against a
// frozen project index.
func BenchmarkCallTargetsOverrides(b *testing.B) {
	var src strings.Builder
	src.WriteString("<?php\nnamespace App;\nabstract class Shape { abstract function area(int $scale): int; function name(): string { return ''; } }\n")
	for i := range 20 {
		fmt.Fprintf(&src, "class S%d extends Shape { function area(int $scale): int { return %d; } }\n", i, i)
		fmt.Fprintf(&src, "class T%d extends S%d { function area(int $scale): int { return 0; } }\n", i, i)
		fmt.Fprintf(&src, "final class U%d extends T%d {}\n", i, i)
	}
	src.WriteString("function probe($x) {}\nfunction run(Shape $s) { probe($s->area(2)); }\n")
	probeFrozen(b, src.String(), func(ctx *analysis.Context, c *syntax.FuncCall) {
		call := CallArgument(c.Args, 0, "")
		callee, _ := ResolveCallee(ctx, call)
		trust := func(string, string) bool { return true }
		if _, targets, ok := CallTargets(ctx, call, callee, trust); !ok || len(targets) != 61 {
			b.Fatalf("targets = %d, %v", len(targets), ok)
		}
		b.ReportAllocs()
		for b.Loop() {
			CallTargets(ctx, call, callee, trust)
		}
	})
}
