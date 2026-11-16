package infer_test

import (
	"fmt"
	"strings"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// TestCorrelatedNullExit covers an early exit taken when two values are
// both null: once a later condition proves one of them null, the other is
// not.
func TestCorrelatedNullExit(t *testing.T) {
	const head = `<?php class A {} function g(): ?A { return null; } function f(?A $a, ?A $b) { `
	cases := []struct {
		name, body, want string
	}{
		{"then branch", `if ($a === null && $b === null) { return; } if ($b === null) { t('v', $a); } }`, `\A`},
		{"null on the left", `if (null === $a && null === $b) { return; } if (null === $b) { t('v', $a); } }`, `\A`},
		{"other order", `if ($b === null && $a === null) { throw new \Exception(); } if ($b === null) { t('v', $a); } }`, `\A`},
		{"else branch", `if ($a === null && $b === null) { return; } if ($b !== null) { } else { t('v', $a); } }`, `\A`},
		{"loose comparison", `if ($a == null && $b == null) { return; } if ($b === null) { t('v', $a); } }`, `\A`},
		{"negated test", `if ($a === null && !($b !== null)) { return; } if ($b === null) { t('v', $a); } }`, `\A`},
		{"null-checked operand", `if (($a === null) and ($b === null)) { return; } if ($b === null) { t('v', $a); } }`, `\A`},
		// An object may be falsy (an empty SimpleXMLElement): a falsy or
		// loosely null $b may still be an A.
		{"loose branch", `if ($a == null && $b == null) { return; } if ($b == null) { t('v', $a); } }`, `\A|null`},
		{"falsy branch", `if (!$a && !$b) { return; } if (!$b) { t('v', $a); } }`, `\A|null`},
		{"falsy exit, null branch", `if (!$a && !$b) { return; } if ($b === null) { t('v', $a); } }`, `\A`},
		{"falsy is not null", `if (!$a && $b === null) { return; } if (!$b) { t('v', $a); } }`, `\A|null`},
		{"early return in a loop", `foreach ([1] as $i) { if ($a === null && $b === null) { continue; } if ($b === null) { t('v', $a); } } }`, `\A`},
		// Not proven.
		{"other value not tested", `if ($a === null && $b === null) { return; } t('v', $a); }`, `\A|null`},
		{"other value non-null", `if ($a === null && $b === null) { return; } if ($b !== null) { t('v', $a); } }`, `\A|null`},
		{"other value written since", `if ($a === null && $b === null) { return; } $b = g(); if ($b === null) { t('v', $a); } }`, `\A|null`},
		{"other value passed to an unknown function", `if ($a === null && $b === null) { return; } unknown_fn($b); if ($b === null) { t('v', $a); } }`, `\A|null`},
		{"other value an int", `if ($a === null && $b === null) { return; } if ($b === 5) { t('v', $a); } }`, `\A|null`},
		{"other value not an object", `if ($a === null && $b === null) { return; } if (!is_object($b)) { t('v', $a); } }`, `\A|null`},
		{"either may be null", `if ($a === null || $b === null) { return; } if ($b === null) { t('v', $a); } }`, `\A`},
		{"not an early exit", `if ($a === null && $b === null) { echo 1; } if ($b === null) { t('v', $a); } }`, `\A|null`},
		{"other test not about null", `if ($a === null && $b instanceof A) { return; } if ($b instanceof A) { t('v', $a); } }`, `\A|null`},
		{"other test reads two values", `if ($a === null && $b === $a) { return; } if ($b === null) { t('v', $a); } }`, `\A|null`},
		{"three values", `if ($a === null && $b === null && g() === null) { return; } if ($b === null) { t('v', $a); } }`, `\A|null`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			native, proven := provenProbe(t, head+c.body)
			if native != c.want || proven != c.want {
				t.Errorf("native %s, proven %s; want %s", native, proven, c.want)
			}
		})
	}
}

// TestCorrelatedExitOfMoreValues covers an early exit taken when several
// values are all null: once later conditions prove all of them but one
// null, the last one is not.
func TestCorrelatedExitOfMoreValues(t *testing.T) {
	const head = `<?php class A {} function g(): ?A { return null; } function f(?A $a, ?A $b, ?A $c, ?A $d) { `
	cases := []struct {
		name, body, want string
	}{
		{"three values", `if ($a === null && $b === null && $c === null) { return; } if ($b === null && $c === null) { t('v', $a); } }`, `\A`},
		{"value in the middle", `if ($b === null && $a === null && $c === null) { return; } if ($c === null) { if ($b === null) { t('v', $a); } } }`, `\A`},
		{"grouped", `if ($b === null and ($c === null && $a === null)) { return; } if ($b === null && $c === null) { t('v', $a); } }`, `\A`},
		{"four values", `if ($a === null && $b === null && $c === null && $d === null) { return; } if ($b === null && $c === null && $d === null) { t('v', $a); } }`, `\A`},
		// Not proven.
		{"one other value not tested", `if ($a === null && $b === null && $c === null) { return; } if ($b === null) { t('v', $a); } }`, `\A|null`},
		{"one other value written since", `if ($a === null && $b === null && $c === null) { return; } $c = g(); if ($b === null && $c === null) { t('v', $a); } }`, `\A|null`},
		{"a call among the tests", `if ($a === null && $b === null && g() === null) { return; } if ($b === null) { t('v', $a); } }`, `\A|null`},
		{"every test proven: unreachable", `if ($b === null && $c === null) { return; } if ($b === null && $c === null) { t('v', $a); } }`, `\A|null`},
		{"a call before the others", `if ($a === null && g() === null && $b === null && $c === null) { return; } if ($b === null && $c === null) { t('v', $a); } }`, `\A|null`},
		{"within the bound", `if ($a === null && $b === null && $c === null && $d === null && $b === null && $c === null && $d === null && $b === null) { return; } if ($b === null && $c === null && $d === null) { t('v', $a); } }`, `\A`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			native, proven := provenProbe(t, head+c.body)
			if native != c.want || proven != c.want {
				t.Errorf("native %s, proven %s; want %s", native, proven, c.want)
			}
		})
	}
}

// TestCorrelatedExitPastTheBound covers exits of more than
// maxCorrelatedTests tests, which are not followed: the native type keeps
// null, and the proven one is unknown wherever following the exit could
// have removed it, so that no check reports what the exit may rule out.
func TestCorrelatedExitPastTheBound(t *testing.T) {
	const head = `<?php class A {} function g(): ?A { return null; } function f(?A $a, ?A $b, ?A $c, ?A $d) { `
	const many = `$b === null && $c === null && $d === null && $b === null && $c === null && $d === null && $b === null && $c === null`
	cases := []struct {
		name, body, proven string
	}{
		{"guarded", `if ($a === null && ` + many + `) { return; } if ($b === null && $c === null && $d === null) { t('v', $a); } }`, `?unknown`},
		{"guarded before the exit", `if ($b === null && $c === null && $d === null) { if ($a === null && ` + many + `) { return; } t('v', $a); } }`, `?unknown`},
		{"exit guard before the exit", `if ($b !== null || $c !== null || $d !== null) { return; } if ($a === null && ` + many + `) { return; } t('v', $a); }`, `?unknown`},
		{"in an else", `if ($a === null && ` + many + `) { return; } if ($b !== null) { } elseif ($c === null && $d === null) { } else { t('v', $a); } }`, `?unknown`},
		{"ternary", `if ($a === null && ` + many + `) { return; } $b === null && $c === null && $d === null ? t('v', $a) : 0; }`, `?unknown`},
		{"switch", `if ($a === null && ` + many + `) { return; } switch ($b === null && $c === null) { case $d === null: t('v', $a); default: } }`, `?unknown`},
		{"switch case", `if ($a === null && ` + many + `) { return; } switch ($b === null && $c === null && $d === null) { case t('v', $a): } }`, `?unknown`},
		{"and", `if ($a === null && ` + many + `) { return; } $b === null && $c === null && $d === null && t('v', $a); }`, `?unknown`},
		{"elseif", `if ($a === null && ` + many + `) { return; } if (g()) { } elseif ($b === null && $c === null && $d === null) { t('v', $a); } }`, `?unknown`},
		{"in a condition", `if ($a === null && ` + many + `) { return; } if ($b === null && $c === null && $d === null && t('v', $a)) { } }`, `?unknown`},
		{"match", `if ($a === null && ` + many + `) { return; } match ($b === null && $c === null) { $d === null => t('v', $a) }; }`, `?unknown`},
		{"loop", `if ($a === null && ` + many + `) { return; } while ($b === null && $c === null) { for (; $d === null;) { t('v', $a); } } }`, `?unknown`},
		{"assertion", `if ($a === null && ` + many + `) { return; } assert($b === null && $c === null && $d === null); t('v', $a); }`, `?unknown`},
		// Nothing on the path tests the other values: the exit cannot
		// have ruled out null for $a.
		{"unguarded", `if ($a === null && ` + many + `) { return; } t('v', $a); }`, `\A|null`},
		{"partly guarded", `if ($a === null && ` + many + `) { return; } if ($b === null && $c === null) { t('v', $a); } }`, `\A|null`},
		{"guard reset since", `if ($b !== null || $c !== null || $d !== null) { return; } $b = g(); if ($a === null && ` + many + `) { return; } t('v', $a); }`, `\A|null`},
		{"a check that does not leave", `if ($a === null && ` + many + `) { return; } if ($b === null && $c === null && $d === null) { g(); } t('v', $a); }`, `\A|null`},
		{"checks that overwrite", `if ($a === null && ` + many + `) { return; } if ($b !== null) { $b = null; } if ($c !== null) { $c = null; } if ($d !== null) { $d = null; } t('v', $a); }`, `?unknown`},
		{"in the if's body", `if ($a === null && ` + many + `) { return; } if (g()) { $b === null && $c === null && $d === null; } else { t('v', $a); } }`, `\A|null`},
		// Following the exit could not have refined $a.
		{"a test that cannot narrow", `if (f($a) && ` + many + `) { return; } if ($b === null && $c === null && $d === null) { t('v', $a); } }`, `\A|null`},
		{"two tests left", `if ($a === null && g() === null && ` + many + `) { return; } t('v', $a); }`, `\A|null`},
		{"no test of the value", `if (` + many + ` && $d === null) { return; } t('v', $a); }`, `\A|null`},
		{"the test left is not about the value", `if (g() === null && ` + many + `) { return; } t('v', $a); }`, `\A|null`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			native, proven := provenProbe(t, head+c.body)
			if native != `\A|null` || proven != c.proven {
				t.Errorf("native %s, proven %s; want \\A|null, %s", native, proven, c.proven)
			}
		})
	}
}

// TestCorrelatedExitPastTheBoundTopLevel covers the same exits in
// top-level code, also inside a namespace block.
func TestCorrelatedExitPastTheBoundTopLevel(t *testing.T) {
	const many = `$b === null && $c === null && $d === null && $b === null && $c === null && $d === null && $b === null && $c === null`
	for _, c := range []struct{ name, src, proven string }{
		{"file", `<?php class A {} $a = $argc ? 1 : null; $b = $argc ? 1 : null; $c = $argc ? 1 : null; $d = $argc ? 1 : null; if ($a === null && ` + many + `) { exit; } if ($b === null && $c === null && $d === null) { t('v', $a); }`, `?unknown`},
		{"namespace", `<?php namespace N { class A {} $a = $argc ? 1 : null; $b = $argc ? 1 : null; $c = $argc ? 1 : null; $d = $argc ? 1 : null; if ($a === null && ` + many + `) { exit; } if ($b === null && $c === null && $d === null) { t('v', $a); } }`, `?unknown`},
		{"file unguarded", `<?php class A {} $a = $argc ? 1 : null; $b = $argc ? 1 : null; $c = $argc ? 1 : null; $d = $argc ? 1 : null; if ($a === null && ` + many + `) { exit; } t('v', $a);`, `int|null`},
		{"namespace unguarded", `<?php namespace N { class A {} $a = $argc ? 1 : null; $b = $argc ? 1 : null; $c = $argc ? 1 : null; $d = $argc ? 1 : null; if ($a === null && ` + many + `) { exit; } t('v', $a); }`, `int|null`},
	} {
		t.Run(c.name, func(t *testing.T) {
			if _, proven := provenProbe(t, c.src); proven != c.proven {
				t.Errorf("proven %s; want %s", proven, c.proven)
			}
		})
	}
}

func TestCorrelatedExitOperandBudget(t *testing.T) {
	for _, count := range []int{8, 9} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			var params, checks, guards []string
			for i := range count {
				name := fmt.Sprintf("$v%d", i)
				params = append(params, "?A "+name)
				checks = append(checks, name+" === null")
				if i > 0 {
					guards = append(guards, name+" === null")
				}
			}
			src := "<?php class A {} function f(" + strings.Join(params, ",") + ") { if (" + strings.Join(checks, " && ") + ") return; if (" + strings.Join(guards, " && ") + ") { t('v', $v0); } }"
			native, proven := provenProbe(t, src)
			wantNative, wantProven := `\A`, `\A`
			if count == 9 {
				wantNative, wantProven = `\A|null`, `?unknown`
			}
			if native != wantNative || proven != wantProven {
				t.Fatalf("native %s proven %s want %s, %s", native, proven, wantNative, wantProven)
			}
		})
	}
}

// BenchmarkCorrelatedExit types every variable of functions whose early
// exit tests n locals for null and whose later uses of the first one are
// guarded by tests of all the others: each use proves n-1 tests.
func BenchmarkCorrelatedExit(b *testing.B) {
	for _, n := range []int{4, 8, 16, 64} {
		b.Run(fmt.Sprint(n), func(b *testing.B) {
			var sb strings.Builder
			sb.WriteString("<?php\nclass A { public function m(): int { return 1; } }\n")
			for f := range 20 {
				var params, checks, guards []string
				for i := range n {
					name := fmt.Sprintf("$v%d", i)
					params = append(params, "?A "+name)
					checks = append(checks, name+" === null")
					if i > 0 {
						guards = append(guards, name+" === null")
					}
				}
				fmt.Fprintf(&sb, "function f%d(%s) {\n    if (%s) { return; }\n", f, strings.Join(params, ", "), strings.Join(checks, " && "))
				for range 5 {
					fmt.Fprintf(&sb, "    if (%s) { $v0->m(); }\n", strings.Join(guards, " && "))
				}
				sb.WriteString("}\n")
			}
			src := []byte(sb.String())
			opt := syntax.Options{Version: phpversion.PHP84}
			b.ReportAllocs()
			for b.Loop() {
				f := syntax.Parse("t.php", src, opt)
				ix := index.New(stubs.Index())
				ix.Add(index.Extract(f))
				env := infer.NewEnv(f, names.New(f), ix, phpversion.PHP84, infer.Options{})
				syntax.InspectFile(f, func(n syntax.Node) bool {
					if v, ok := n.(*syntax.Variable); ok {
						env.ProvenTypeOf(v)
					}
					return true
				})
			}
		})
	}
}
