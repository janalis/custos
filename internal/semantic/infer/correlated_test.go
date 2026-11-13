package infer_test

import (
	"fmt"
	"strings"
	"testing"
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
		{"past the bound", `if ($a === null && $b === null && $c === null && $d === null && $b === null && $c === null && $d === null && $b === null && $c === null && $d === null) { return; } if ($b === null && $c === null && $d === null) { t('v', $a); } }`, `\A|null`},
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
			want := `\A`
			if count == 9 {
				want = `\A|null`
			}
			if native != want || proven != want {
				t.Fatalf("native %s proven %s want %s", native, proven, want)
			}
		})
	}
}
