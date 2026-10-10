// Package splobjectstorageindirectarraymutation implements the native SplObjectStorageIndirectArrayMutation inspection.
package splobjectstorageindirectarraymutation

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Assign the modified storage value back."

type rule struct{}

func New() analysis.Rule { return rule{} }
func (rule) ID() string  { return "SplObjectStorageIndirectArrayMutation" }
func (rule) Semantic()   {}
func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KAssign, syntax.KIncDec, syntax.KUnset}
}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	var targets []syntax.Expr
	switch x := n.(type) {
	case *syntax.Assign:
		targets = []syntax.Expr{x.Var}
	case *syntax.IncDec:
		targets = []syntax.Expr{x.Var}
	case *syntax.Unset:
		targets = x.Vars
	}
	for _, target := range targets {
		nested, ok := syntax.UnwrapParens(target).(*syntax.ArrayDimFetch)
		if !ok {
			continue
		}
		slot, ok := nested.Var.(*syntax.ArrayDimFetch)
		if !ok {
			continue
		}
		if semanticquery.NativeConstruction(ctx, slot.Var, "SplObjectStorage") == nil {
			continue
		}
		if semanticquery.NativeStoredArray(ctx, slot, n) {
			ctx.ReportNode(n, message)
			return
		}
	}
}
