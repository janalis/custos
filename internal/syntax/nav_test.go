package syntax

import (
	"testing"

	"custos/internal/phpver"
)

func TestNavHelpers(t *testing.T) {
	f := Parse("t.php", []byte(`<?php class C { function m($a, $b = null) { return fn($x) => ((($x))); } }`), Options{Version: phpver.Max})
	var v *Variable
	var arrow *ArrowFunction
	var method *Method
	InspectFile(f, func(n Node) bool {
		switch x := n.(type) {
		case *Variable:
			if x.Name == "x" && v == nil {
				if _, ok := x.Parent().(*Paren); ok {
					v = x
				}
			}
		case *ArrowFunction:
			arrow = x
		case *Method:
			method = x
		}
		return true
	})
	if v == nil || arrow == nil || method == nil {
		t.Fatal("nodes not found")
	}
	if EnclosingFuncLike(v) != Node(arrow) || EnclosingFuncLike(arrow) != Node(method) {
		t.Error("EnclosingFuncLike")
	}
	if c := EnclosingClass(v); c == nil || c.Name.Value != "C" {
		t.Error("EnclosingClass")
	}
	if UnwrapParens(arrow.Expr) != Expr(v) {
		t.Error("UnwrapParens")
	}
	if !IsFuncLike(arrow) || IsFuncLike(v) {
		t.Error("IsFuncLike")
	}
	if len(FuncLikeParams(method)) != 2 || len(FuncLikeParams(arrow)) != 1 || FuncLikeBody(arrow) != nil || FuncLikeBody(method) == nil {
		t.Error("FuncLikeParams/FuncLikeBody")
	}
	if !IsNullConst(method.Params[1].Default) || IsNullConst(arrow.Expr) {
		t.Error("IsNullConst")
	}
}
