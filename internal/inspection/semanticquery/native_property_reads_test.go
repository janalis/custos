package semanticquery

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestNativePropertyUnchanged(t *testing.T) {
	const classes = `final class F { public $p; public $q; } class O { public $p; } class M { public $p; function __get($n) {} } final class H { public $p { get => 1; } } `
	for _, tc := range []struct {
		name, src string
		want      bool
	}{
		{"file level", `$o = new F(); $first = $o->p; probe($o->p);`, true},
		{"nothing between", `function t(F $o) { $first = $o->p; probe($o->p); }`, true},
		{"private property", `class P { private $p; function m() { $first = $this->p; probe($this->p); } }`, true},
		{"pure call between", `function t(F $o) { $first = $o->p; strlen('x'); probe($o->p); }`, true},
		{"other property written", `function t(F $o) { $first = $o->p; $o->q = 1; probe($o->p); }`, true},
		{"open class", `function t(O $o) { $first = $o->p; probe($o->p); }`, false},
		{"magic read", `function t(M $o) { $first = $o->p; probe($o->p); }`, false},
		{"hooked property", `function t(H $o) { $first = $o->p; probe($o->p); }`, false},
		{"unknown receiver", `function t($o) { $first = $o->p; probe($o->p); }`, false},
		{"null-safe read", `function t(F $o) { $first = $o?->p; probe($o?->p); }`, false},
		{"other receiver", `function t(F $o, F $r) { $first = $o->p; probe($r->p); }`, false},
		{"dynamic name", `function t(F $o, $n) { $first = $o->$n; probe($o->$n); }`, false},
		{"project call between", `function t(F $o) { $first = $o->p; work(); probe($o->p); }`, false},
		{"method call between", `function t(F $o) { $first = $o->p; $o->m(); probe($o->p); }`, false},
		{"construction between", `function t(F $o) { $first = $o->p; new F(); probe($o->p); }`, false},
		{"evaluated code between", `function t(F $o, $c) { $first = $o->p; eval($c); probe($o->p); }`, false},
		{"written through another receiver", `function t(F $o, F $r) { $first = $o->p; $r->p = 1; probe($o->p); }`, false},
		{"written through a dynamic name", `function t(F $o, $n) { $first = $o->p; $o->$n = 1; probe($o->p); }`, false},
		{"reference taken", `function t(F $o) { $first = $o->p; $x = &$o->p; probe($o->p); }`, false},
		{"incremented", `function t(F $o) { $first = $o->p; $o->p++; probe($o->p); }`, false},
		{"unset", `function t(F $o) { $first = $o->p; unset($o->p); probe($o->p); }`, false},
		{"foreach target", `function t(F $o, $l) { $first = $o->p; foreach ($l as $o->p) {} probe($o->p); }`, false},
		{"by-reference argument", `function t(F $o) { $first = $o->p; preg_match('/a/', 'a', $o->p); probe($o->p); }`, false},
		{"receiver rebound", `function t(F $o, F $r) { $first = $o->p; $o = $r; probe($o->p); }`, false},
		{"closure declared between", `function t(F $o) { $first = $o->p; $f = function () use ($o) { $o->p = 1; }; probe($o->p); }`, true},
		{"written later in a loop", `function t(F $o, $c) { $first = $o->p; while ($c) { probe($o->p); $o->p = 1; } }`, false},
		{"loop around both reads", `function t(F $o, $c) { while ($c) { $first = $o->p; probe($o->p); $o->p = 1; } }`, true},
		{"read before", `function t(F $o) { probe($o->p); $first = $o->p; }`, false},
		{"other scope", `function t(F $o) { $first = $o->p; $f = function (F $o) { probe($o->p); }; }`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			probeNative(t, classes+tc.src, func(ctx *analysis.Context, probe *syntax.FuncCall) {
				var earlier syntax.Expr
				syntax.InspectFile(ctx.File, func(n syntax.Node) bool {
					if assign, ok := n.(*syntax.Assign); ok {
						if target, plain := assign.Var.(*syntax.Variable); plain && target.Name == "first" && (syntax.EnclosingVariableScope(assign) == syntax.EnclosingVariableScope(probe) || tc.name == "other scope") {
							earlier = assign.Value
						}
					}
					return true
				})
				if got := NativePropertyUnchanged(ctx, earlier, CallArgument(probe.Args, 0, "")); got != tc.want {
					t.Fatalf("unchanged=%v", got)
				}
			})
		})
	}
}
