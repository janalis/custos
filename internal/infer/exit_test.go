package infer_test

import (
	"testing"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/syntax"
	"custos/internal/types"
)

func TestExitCallableInference(t *testing.T) {
	for _, expression := range []string{`exit(...)`, `die(...)`, `exit`, `die()`, `exit(status: 1)`, `(exit(...))(1)`, `(die(...))('done')`} {
		t.Run(expression, func(t *testing.T) {
			f := syntax.Parse("exit.php", []byte("<?php "+expression+";"), syntax.Options{Version: phpver.PHP84})
			if len(f.Errors) != 0 {
				t.Fatal(f.Errors)
			}
			env := infer.NewEnv(f, names.New(f), index.New(nil), phpver.PHP84)
			got := env.TypeOf(f.Stmts[0].(*syntax.ExprStmt).Expr)
			want := "never"
			if expression == `exit(...)` || expression == `die(...)` {
				want = `\Closure`
				if ret := types.CallableReturn(got).String(); ret != "never" {
					t.Errorf("callable return = %s", ret)
				}
			}
			if got.String() != want {
				t.Errorf("got %s, want %s", got, want)
			}
			if native := env.Native().TypeOf(f.Stmts[0].(*syntax.ExprStmt).Expr); native.DocString() != got.DocString() {
				t.Errorf("native type = %s, want %s", native.DocString(), got.DocString())
			}
		})
	}
	checkVer(t, phpver.PHP84, false, `<?php function f(?string $s) {
 if ($s === null) { exit(...); }
 t('nullable', $s);
 $x = 1;
 exit(...);
 t('reachable', $x);
}`, map[string]string{"nullable": "null|string", "reachable": "int"})
}

func TestImmediateExitInvocationFlow(t *testing.T) {
	checkVer(t, phpver.PHP84, false, `<?php
function guard(?string $s) {
 if ($s === null) { ((exit(...))(1)); }
 t('narrowed', $s);
}
function callableGuard(?string $s) {
 if ($s === null) { (exit(...))(...); }
 t('nullable', $s);
}
function exitGenerator() { yield 1; ((exit(...))(1)); }
function dieGenerator() { yield 1; (die(...))('done'); }
function callableGenerator() { yield 1; (exit(...))(...); }
t('exitCompletion', exitGenerator()->getReturn());
t('dieCompletion', dieGenerator()->getReturn());
t('callableCompletion', callableGenerator()->getReturn());
`, map[string]string{"narrowed": "string", "nullable": "null|string", "exitCompletion": "never", "dieCompletion": "never", "callableCompletion": "null"})
}

func TestExitAssignmentFlow(t *testing.T) {
	checkVer(t, phpver.PHP84, false, `<?php
function exiting($flag) {
 if ($flag) { $value = 'assigned'; } else { ((exit(...))(1)); }
 t('assigned', $value);
}
function creating($flag) {
 if ($flag) { $value = 'assigned'; } else { (exit(...))(...); }
 t('possiblyUndefined', $value);
}
`, map[string]string{"assigned": "string", "possiblyUndefined": "null|string"})
}

func BenchmarkExitInference(b *testing.B) {
	for _, expression := range []string{`exit()`, `exit(...)`} {
		b.Run(expression, func(b *testing.B) {
			f := syntax.Parse("exit.php", []byte("<?php "+expression+";"), syntax.Options{Version: phpver.PHP84})
			ix, resolver := index.New(nil), names.New(f)
			expr := f.Stmts[0].(*syntax.ExprStmt).Expr
			b.ReportAllocs()
			for b.Loop() {
				infer.NewEnv(f, resolver, ix, phpver.PHP84).TypeOf(expr)
			}
		})
	}
}
