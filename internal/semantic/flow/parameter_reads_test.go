package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestParameterRead(t *testing.T) {
	for _, tc := range []struct {
		body        string
		read, known bool
	}{
		{`echo $x;`, true, true},
		{`$x = 1; echo $x;`, false, true},
		{`if ($flag) { $x = 1; } echo $x;`, true, true},
		{`if ($flag) { $x = 1; } else { $x = 2; } echo $x;`, false, true},
		{`while ($flag) { $x = 1; } echo $x;`, true, true},
		{`do { $x = 1; } while ($flag); echo $x;`, false, true},
		{`$x = $x + 1; echo $x;`, true, true},
		{`$x = unknown(); echo $x;`, false, false},
		{`eval($code); echo $x;`, false, false},
		{`$saved = get_defined_vars(); $x = 1; echo $x;`, false, false},
		{`$saved = compact('x'); $x = 1; echo $x;`, false, false},
		{`$saved = func_get_args(); $x = 1; echo $x;`, false, false},
		{`$saved = func_get_arg(0); $x = 1; echo $x;`, false, false},
		{`$saved = $$name; $x = 1; echo $x;`, false, false},
		{`${$flag} = 1; $x = 1; echo $x;`, false, true},
	} {
		t.Run(tc.body, func(t *testing.T) {
			e, f := environment(t, `function f($x, $flag) { `+tc.body+` }`, nil)
			fn := f.Stmts[0].(*syntax.Function)
			read, known := e.ParameterRead(fn, 0)
			if read != tc.read || known != tc.known {
				t.Fatalf("got %v, %v; want %v, %v", read, known, tc.read, tc.known)
			}
			for _, index := range []int{-1, 2} {
				if _, ok := e.ParameterRead(fn, index); ok {
					t.Fatal("invalid parameter")
				}
			}
		})
	}
}

func BenchmarkParameterRead(b *testing.B) {
	e, f := environment(b, `function f($x, $flag) { if ($flag) { $x = 1; } else { $x = 2; } echo $x; }`, nil)
	fn := f.Stmts[0].(*syntax.Function)
	e.ParameterRead(fn, 0)
	b.ReportAllocs()
	for b.Loop() {
		e.ParameterRead(fn, 0)
	}
}

func TestParameterReadInvalidDeclarations(t *testing.T) {
	e, f := environment(t, `function f(&$x) { echo $x; }`, nil)
	if _, known := e.ParameterRead(f.Stmts[0], 0); known {
		t.Fatal("reference parameter")
	}
}

func TestParameterReadCaptures(t *testing.T) {
	for _, body := range []string{`$cb = fn() => $x; $x = 1;`, `$cb = function() use ($x) {}; $x = 1;`, `class Nested {} function other() {} $cb = function() use (&$x) {};`} {
		e, f := environment(t, `function f($x) { `+body+` }`, nil)
		if read, known := e.ParameterRead(f.Stmts[0], 0); !read && known {
			t.Fatal("capture proved unused")
		}
	}
}
