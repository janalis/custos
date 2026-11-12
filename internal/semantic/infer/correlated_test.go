package infer_test

import "testing"

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
