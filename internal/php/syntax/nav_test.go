package syntax

import (
	"testing"

	phpversion "custos/internal/php/version"
)

func TestNavHelpers(t *testing.T) {
	f := Parse("t.php", []byte(`<?php class C { function m($a, $b = null) { return fn($x) => ((($x))); } }`), Options{Version: phpversion.Max})
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

// The constructor arguments of an anonymous class belong to the scope that
// creates it; its body belongs to the anonymous class.
func TestEnclosingClassAnonymousArguments(t *testing.T) {
	f := Parse("t.php", []byte(`<?php class Outer { function m() { return new class($this->a) { function n() { return $this->b; } }; } }`), Options{Version: phpversion.Max})
	fetches := map[string]*PropertyFetch{}
	InspectFile(f, func(n Node) bool {
		if p, ok := n.(*PropertyFetch); ok {
			fetches[p.Name.(*Identifier).Value] = p
		}
		return true
	})
	if c := EnclosingClass(fetches["a"]); c == nil || c.Name == nil || c.Name.Value != "Outer" {
		t.Errorf("argument: got %v, want Outer", c)
	}
	if c := EnclosingClass(fetches["b"]); c == nil || c.Name != nil {
		t.Errorf("body: got %v, want the anonymous class", c)
	}
}
