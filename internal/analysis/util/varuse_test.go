package util

import (
	"strings"
	"testing"

	"custos/internal/syntax"
)

// describe renders accesses as e.g. "w r e c" (write, read, elem write, compound).
func describe(as []VarAccess) string {
	var parts []string
	for _, a := range as {
		switch {
		case a.Write && a.Compound:
			parts = append(parts, "c")
		case a.Write:
			parts = append(parts, "w")
		case a.ElemWrite:
			parts = append(parts, "e")
		default:
			parts = append(parts, "r")
		}
	}
	return strings.Join(parts, " ")
}

func TestVarAccesses(t *testing.T) {
	cases := []struct{ body, want string }{
		{`$v = 1; return $v;`, "w r"},
		{`$v = $v + 1;`, "w r"},
		{`$v .= 'x'; $v++;`, "c c"},
		{`[$a, $v] = f(); list('k' => $v) = g();`, "w w"},
		{`foreach ($v as $k => $v) {}`, "r w"},
		{`foreach ($xs as [$v, $w]) {}`, "w"},
		{`global $v; static $v = 1;`, "w w"},
		{`try {} catch (E $v) { echo $v; }`, "w r"},
		{`unset($v, $v['k']);`, "w r"},
		{`$v[] = 1; $v['a']['b'] = 2; $v->p = 3;`, "e e e"},
		{`$v->m(); isset($v); echo "x $v";`, "r r r"},
		{`$f = function () use ($v) { $v = 2; return $v; };`, "r"},
		{`$f = fn() => $v + 1; $g = fn($v) => $v;`, "r"},
		{`$f = fn() => $v = 3;`, ""},
		{`function inner() { $v = 1; } class K { function m() { $v = 2; } }`, ""},
		{`$$v = 1; ${'v'} = 2;`, "r"},
		{`$w = &$v;`, "r"},
	}
	for _, c := range cases {
		f := parse(t, "<?php function scope($p) { "+c.body+" }")
		fn := f.Stmts[0].(*syntax.Function)
		if got := describe(VarAccesses(f, fn, "v")); got != c.want {
			t.Errorf("%s\n got %q, want %q", c.body, got, c.want)
		}
	}
}

func TestVarAccessesFileScopeAndCount(t *testing.T) {
	f := parse(t, "<?php $v = 1; echo $v; function g() { return $v; } $v = 2;")
	as := VarAccesses(f, nil, "v")
	if got := describe(as); got != "w r w" {
		t.Fatalf("file scope: %q", got)
	}
	if r, w := CountVarAccesses(as); r != 1 || w != 2 {
		t.Fatalf("count = %d reads, %d writes", r, w)
	}
	if _, ok := as[0].By.(*syntax.Assign); !ok {
		t.Fatal("By must point at the assignment")
	}
	f = parse(t, "<?php $f = fn() => $v * 2;")
	af := f.Stmts[0].(*syntax.ExprStmt).Expr.(*syntax.Assign).Value
	if got := describe(VarAccesses(f, af, "v")); got != "r" {
		t.Fatalf("arrow scope: %q", got)
	}
}
