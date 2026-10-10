// Package bcmathfloatoperand implements the native BcMathFloatOperand inspection.
package bcmathfloatoperand

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply decimal strings before BCMath arithmetic."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "BcMathFloatOperand" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	for _, operand := range semanticquery.ExpansionBcOperands(ctx, c) {
		if operand != nil && ctx.Types().Native().TypeOf(operand).OnlyOf("float") {
			ctx.ReportNode(c, message)
			return
		}
	}
}
