package util

import (
	"testing"

	"custos/internal/syntax"
)

func TestEquivalent(t *testing.T) {
	cases := []struct {
		src  string
		want bool
	}{
		{`<?php f($a, $a);`, true},
		{`<?php f($a['x'], $a[ /* c */ 'x' ]);`, true},
		{`<?php f($a, ($a));`, false},
		{`<?php f($a, $b);`, false},
		{`<?php f($a->b, $a->bc);`, false},
	}
	for _, c := range cases {
		f := parse(t, c.src)
		call := firstExpr(t, f).(*syntax.FuncCall)
		a := call.Args.Args[0].(*syntax.Arg).Value
		b := call.Args.Args[1].(*syntax.Arg).Value
		if got := Equivalent(f, a, b); got != c.want {
			t.Errorf("%s: got %v", c.src, got)
		}
	}
}

func TestFunctionLike(t *testing.T) {
	f := parse(t, `<?php function g($p) { $x = fn($q) => $q; }`)
	var arrow syntax.Node
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if _, ok := n.(*syntax.Variable); ok && arrow == nil {
			if v := n.(*syntax.Variable); v.Name == "x" {
				arrow = n
			}
		}
		return true
	})
	fn := EnclosingFuncLike(arrow)
	if fn == nil || len(FuncLikeParams(fn)) != 1 || FuncLikeParams(fn)[0].Var.Name != "p" {
		t.Fatalf("unexpected enclosing function %v", fn)
	}
}
