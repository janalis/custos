package flowquery

import (
	"testing"

	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
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
		f := Parse(t, c.src)
		var arg syntax.Expr
		syntax.InspectFile(f, func(n syntax.Node) bool {
			if call, ok := n.(*syntax.FuncCall); ok && astquery.CallLastName(call) == "g" {
				arg = call.Args.Args[0].(*syntax.Arg).Value
			}
			return true
		})
		got := ""
		if lit := SingleStringLiteral(f, arg); lit != nil {
			got = Text(f, lit)
		}
		if got != c.want {
			t.Errorf("%s: got %q, want %q", c.src, got, c.want)
		}
	}
}

func TestSingleStringLiteralNumber(t *testing.T) {
	f := Parse(t, `<?php 1;`)
	if SingleStringLiteral(f, FirstExpr(t, f)) != nil {
		t.Fatal("a number literal is not a string")
	}
}

func TestConstantStringWithReferences(t *testing.T) {
	for _, tc := range []struct {
		expr  string
		found bool
	}{
		{"<<<'END'\n/a/\nEND", true},
		{"<<<END\n/a/\nEND", true},
		{"<<<END\n/$x/\nEND", false},
		{`'a'`, true},
		{`123`, false},
		{`$flag ? 'a' : 'b'`, false},
		{`$flag ? 'a' : 123`, true},
	} {
		f := Parse(t, "<?php function f() { probe("+tc.expr+"); }")
		if got := ConstantStringWithReferences(f, ProbeArg(f), nil); (got != nil) != tc.found {
			t.Fatalf("%s: %v", tc.expr, got)
		}
	}
	f := Parse(t, `<?php function f() { $x = 'a'; $x .= 'b'; probe($x); }`)
	if ConstantStringWithReferences(f, ProbeArg(f), nil) != nil {
		t.Fatal("unstable string")
	}
}
