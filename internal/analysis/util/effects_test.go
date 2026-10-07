package util

import "testing"

func TestMayHaveSideEffects(t *testing.T) {
	cases := map[string]bool{
		`$a`:                false,
		`$a['k']->p + 1`:    false,
		`isset($a) && !$b`:  false,
		`fn() => f()`:       false,
		`f()`:               true,
		`$o->m()`:           true,
		`A::m()`:            true,
		`new A`:             true,
		`$a = 1`:            true,
		`$a++`:              true,
		`clone $a`:          true,
		"`ls`":              true,
		`$x ?: g($y)`:       true,
		`$a |> strlen(...)`: true,
	}
	for src, want := range cases {
		e := firstExpr(t, parse(t, "<?php "+src+";"))
		if got := MayHaveSideEffects(e); got != want {
			t.Errorf("%s: got %v, want %v", src, got, want)
		}
	}
}
