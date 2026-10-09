package semanticquery

import (
	"testing"

	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
)

func TestCallHelperEdges(t *testing.T) {
	f := Parse(t, `<?php $a; f(...);`)
	if astquery.CallLastName(FirstExpr(t, f)) != "" {
		t.Fatal("CallLastName of a non-call")
	}
	if _, ok := astquery.CallArgValues(&syntax.FuncCall{}); ok {
		t.Fatal("call without argument list")
	}
	if astquery.ArgCount(&syntax.FuncCall{}) != 0 {
		t.Fatal("ArgCount without argument list")
	}
	call := f.Stmts[1].(*syntax.ExprStmt).Expr.(*syntax.FuncCall)
	if ArgBindsByRef(call.Args, []index.Param{{Name: "$x", ByRef: true}}, true) {
		t.Fatal("the first-class callable placeholder binds nothing")
	}
}
