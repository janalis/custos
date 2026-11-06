package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestPureNextBudgetsAndUnknowns(t *testing.T) {
	e, _ := environment(t, ``, nil)
	one := pureConstant(constantScalar{k: 'i', i: 1})
	unknown := &pureExpr{kind: pureSlot, slot: 0}
	for _, expr := range []*pureExpr{
		{kind: pureNonNull, left: unknown, right: one},
		{kind: pureCount, left: &pureExpr{kind: pureArray}, right: &pureExpr{kind: pureSlot, slot: 1}},
	} {
		if _, ok := e.evalPure(expr, []Value{{}}, &scopeResult{}, MaxDepth); ok {
			t.Fatal("unknown input was proven")
		}
	}
	if (&pureCompiler{env: e}).finiteNative(&syntax.FuncCall{}, "count", nil) != nil {
		t.Fatal("missing builtin roles")
	}
	array := &pureExpr{kind: pureArray, args: []*pureExpr{pureConstant(constantScalar{k: 's', s: "0"}), one}}
	member := &pureExpr{kind: pureMember, left: one, right: array, other: pureConstant(constantScalar{k: 'b', b: true})}
	if _, ok := e.evalPure(member, nil, &scopeResult{operations: MaxTransfers - 5}, MaxDepth); ok {
		t.Fatal("membership bypassed work limit")
	}
}

func TestPureAnonymousInstanceRemainsUnknown(t *testing.T) {
	e, file := environment(t, `$k=new class{function wrap($x){return $this->leaf($x);}};`, nil)
	var call *syntax.MethodCall
	syntax.InspectFile(file, func(node syntax.Node) bool {
		if n, ok := node.(*syntax.MethodCall); ok {
			call = n
		}
		return true
	})
	if (&pureCompiler{env: e}).instanceCall(call) != nil {
		t.Fatal("anonymous receiver established a contract")
	}
}
