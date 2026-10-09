package syntax

import (
	"testing"

	phpversion "custos/internal/php/version"
)

func TestVariableScopes(t *testing.T) {
	f := Parse("t.php", []byte(`<?php
class C {
 public int $x {
  get { $f = fn() => $local; return $f(); }
  set(int $incoming) { $local = $incoming; }
 }
 public function __construct(public int $y { set => $value; }) {}
}
outside();
`), Options{Version: phpversion.PHP84})
	if len(f.Errors) != 0 {
		t.Fatal(f.Errors)
	}
	var hooks []*PropertyHook
	var arrow *ArrowFunction
	InspectFile(f, func(n Node) bool {
		switch n := n.(type) {
		case *PropertyHook:
			hooks = append(hooks, n)
			if !IsVariableScope(n) || IsFuncLike(n) || FuncLikeBody(n) != nil || FuncLikeParams(n) != nil {
				t.Fatal("hook changed function-like contracts")
			}
			if VariableScopeBody(n) != n.Body || len(VariableScopeParams(n)) != len(n.Params) {
				t.Fatal("hook body/params")
			}
		case *ArrowFunction:
			arrow = n
			if !IsVariableScope(n) || VariableScopeBody(n) != n.Expr || VariableScopeParams(n) != nil {
				t.Fatal("arrow body/params")
			}
		}
		return true
	})
	if len(hooks) != 3 || arrow == nil || EnclosingVariableScope(arrow) != hooks[0] || EnclosingVariableScope(arrow.Expr) != arrow {
		t.Fatal("nested hook scopes")
	}
	if EnclosingVariableScope(hooks[0]) != nil {
		t.Fatal("class is not a variable scope")
	}
	if _, ok := EnclosingVariableScope(hooks[2]).(*Method); !ok {
		t.Fatal("promoted hook's enclosing declaration")
	}
	if EnclosingVariableScope(f.Stmts[len(f.Stmts)-1]) != nil || IsVariableScope(f.Stmts[len(f.Stmts)-1]) || VariableScopeBody(f.Stmts[len(f.Stmts)-1]) != nil || VariableScopeParams(f.Stmts[len(f.Stmts)-1]) != nil {
		t.Fatal("top-level scope")
	}
	fn := &Function{Body: &Block{}, Params: []*Param{{}}}
	if VariableScopeBody(fn) != fn.Body || len(VariableScopeParams(fn)) != 1 {
		t.Fatal("function body/params")
	}
}
