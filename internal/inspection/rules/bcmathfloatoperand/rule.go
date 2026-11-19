// Package bcmathfloatoperand implements the native BcMathFloatOperand inspection.
package bcmathfloatoperand

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/flowquery"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Supply decimal strings before BCMath arithmetic."

type rule struct{}

// New returns the BcMathFloatOperand inspection.
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
	for _, operand := range semanticquery.BCMathOperands(ctx, c) {
		if operand != nil && ctx.Types().Native().TypeOf(operand).OnlyOf("float") {
			ctx.ReportNode(c, message)
			return
		}
	}
}
