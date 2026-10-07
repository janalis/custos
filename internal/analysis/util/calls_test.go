package util

import (
	"testing"

	"custos/internal/syntax"
)

func TestCallHelpers(t *testing.T) {
	f := parse(t, `<?php \Ns\foo($a, 'b');`)
	call := firstExpr(t, f).(*syntax.FuncCall)
	q, n, ok := CallName(call)
	if !ok || q != `\Ns\` || n != "foo" {
		t.Fatalf("CallName = %q %q %v", q, n, ok)
	}
	if CallLastName(call) != "foo" {
		t.Fatal("CallLastName")
	}
	args, ok := CallArgValues(call)
	if !ok || len(args) != 2 || text(f, args[1]) != "'b'" {
		t.Fatalf("CallArgValues = %v %v", args, ok)
	}
	if ParentFuncCall(args[0]) != call {
		t.Fatal("ParentFuncCall")
	}
	if ParentFuncCall(call) != nil {
		t.Fatal("ParentFuncCall of statement expr")
	}
	f = parse(t, `<?php f(...$a);`)
	if _, ok := CallArgValues(firstExpr(t, f).(*syntax.FuncCall)); ok {
		t.Fatal("spread must fail")
	}
	f = parse(t, `<?php $f();`)
	if _, _, ok := CallName(firstExpr(t, f).(*syntax.FuncCall)); ok {
		t.Fatal("dynamic call")
	}
}

func TestBoolConst(t *testing.T) {
	for src, want := range map[string][2]bool{
		`<?php TRUE;`:   {true, true},
		`<?php \false;`: {false, true},
		`<?php null;`:   {false, false},
		`<?php $true;`:  {false, false},
	} {
		f := parse(t, src)
		v, ok := BoolConst(firstExpr(t, f))
		if v != want[0] || ok != want[1] {
			t.Errorf("%s: got %v %v", src, v, ok)
		}
	}
}
