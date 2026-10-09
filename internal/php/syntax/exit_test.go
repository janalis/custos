package syntax

import (
	"testing"

	phpversion "custos/internal/php/version"
)

func TestExitCallableTermination(t *testing.T) {
	for _, tc := range []struct {
		src     string
		invokes bool
	}{
		{`exit;`, true},
		{`die;`, true},
		{`exit();`, true},
		{`die(1);`, true},
		{`exit(status: 1);`, true},
		{`exit(...$args);`, true},
		{`exit(1, 2);`, true},
		{`exit(...);`, false},
		{`die(...);`, false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			f := parse(t, "<?php "+tc.src+" next();", phpversion.PHP84)
			if len(f.Errors) != 0 {
				t.Fatal(f.Errors)
			}
			n := firstExpr(t, f).(*Exit)
			if got := ExitInvokes(n); got != tc.invokes {
				t.Errorf("invokes = %v", got)
			}
			if got := Terminates(f.Stmts[0]); got != tc.invokes {
				t.Errorf("terminates = %v", got)
			}
			want := len(f.Stmts)
			if tc.invokes {
				want = 0
			}
			if got := FirstTerminating(&Block{Stmts: f.Stmts}); got != want {
				t.Errorf("first terminating = %d, want %d", got, want)
			}
		})
	}
	f := parse(t, `<?php (exit(...));`, phpversion.PHP84)
	if Terminates(f.Stmts[0]) {
		t.Fatal("parenthesized callable terminates")
	}
}

func BenchmarkExitTermination(b *testing.B) {
	for _, src := range []string{`exit();`, `exit(...);`} {
		b.Run(src, func(b *testing.B) {
			f := Parse("exit.php", []byte("<?php "+src), Options{Version: phpversion.PHP84})
			b.ReportAllocs()
			for b.Loop() {
				Terminates(f.Stmts[0])
			}
		})
	}
}

func TestImmediateExitInvocation(t *testing.T) {
	for _, tc := range []struct {
		src     string
		invokes bool
	}{
		{`exit();`, true},
		{`exit(...);`, false},
		{`(exit(...))(1);`, true},
		{`((die(...)))('done');`, true},
		{`((exit(...))(status: 1));`, true},
		{`(exit(...))(...);`, false},
		{`(die(...))(...$args);`, true},
		{`f();`, false},
		{`(exit())(1);`, false},
		{`$callback(1);`, false},
	} {
		t.Run(tc.src, func(t *testing.T) {
			f := parse(t, "<?php "+tc.src, phpversion.PHP84)
			if len(f.Errors) != 0 {
				t.Fatal(f.Errors)
			}
			expr := firstExpr(t, f)
			if got := ExitInvocation(expr); got != tc.invokes {
				t.Errorf("invocation = %v", got)
			}
			if got := Terminates(f.Stmts[0]); got != tc.invokes {
				t.Errorf("terminates = %v", got)
			}
		})
	}
	if ExitInvocation(nil) {
		t.Fatal("nil invokes exit")
	}
}
