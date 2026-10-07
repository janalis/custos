package util

import (
	"testing"

	"custos/internal/syntax"
)

func TestSingleStringLiteral(t *testing.T) {
	cases := []struct {
		src, want string
	}{
		{`<?php function f() { $m = 'w'; g($m); }`, `'w'`},
		{`<?php function f($c) { $m = $c ? 'a' : 'b'; g($m); }`, ``},
		{`<?php function f($c) { $m = $c ? 'a' : 1; g($m); }`, `'a'`},
		{`<?php g('x');`, `'x'`},
		{`<?php g($top);`, ``},
		{`<?php function f($c) { $m = 'a'; $n = 1; $n++; $m = $c ? 3 : $n; g($m); }`, ``},
		{`<?php function f() { $m = '/a/'; $m .= 'i'; g($m); }`, ``},
	}
	for _, c := range cases {
		f := parse(t, c.src)
		var arg syntax.Expr
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if call, ok := n.(*syntax.FuncCall); ok && CallLastName(call) == "g" {
				arg = call.Args.Args[0].(*syntax.Arg).Value
			}
			return true
		})
		got := ""
		if lit := SingleStringLiteral(f, arg); lit != nil {
			got = text(f, lit)
		}
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.src, got, c.want)
		}
	}
}

func TestSingleStringLiteralNumber(t *testing.T) {
	f := parse(t, `<?php 1;`)
	if SingleStringLiteral(f, firstExpr(t, f)) != nil {
		t.Fatal("a number literal is not a string")
	}
}
