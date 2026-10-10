// Package seeksuccesscheckedastruthy implements the native SeekSuccessCheckedAsTruthy inspection.
package seeksuccesscheckedastruthy

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare the seek result with zero."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SeekSuccessCheckedAsTruthy" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if semanticquery.NativeBuiltin(ctx, c, "fseek") && semanticquery.NativeConditionUse(ctx, c, true) != nil {
		ctx.ReportNode(c, message)
	}
}
