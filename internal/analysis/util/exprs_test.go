package util

import (
	"testing"

	"custos/internal/syntax"
)

func TestFuncNameHelpers(t *testing.T) {
	f := parse(t, `<?php \Foo\array_search($a, ...$b);`)
	call := firstExpr(t, f).(*syntax.FuncCall)
	name, part, ok := FuncNamePart(call)
	if !ok || part != "array_search" {
		t.Fatalf("FuncNamePart = %q, %v", part, ok)
	}
	if got := spanText(f, NamePartSpan(name)); got != "array_search" {
		t.Fatalf("NamePartSpan = %q", got)
	}
	if _, ok := IsFuncNamed(call, "array_search"); !ok {
		t.Fatal("IsFuncNamed: want match")
	}
	if _, ok := IsFuncNamed(call, "Array_Search"); ok {
		t.Fatal("IsFuncNamed: case-sensitive")
	}
	if n := ArgCount(call); n != 2 {
		t.Fatalf("ArgCount = %d", n)
	}
	dyn := firstExpr(t, parse(t, `<?php $f();`)).(*syntax.FuncCall)
	if _, _, ok := FuncNamePart(dyn); ok || ArgCount(dyn) != 0 {
		t.Fatal("dynamic call")
	}
}
