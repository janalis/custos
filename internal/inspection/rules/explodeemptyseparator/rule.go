// Package explodeemptyseparator implements the ExplodeEmptySeparator inspection.
package explodeemptyseparator

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ExplodeEmptySeparator" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Supply a nonempty separator."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "explode")
	if call == nil {
		return
	}
	separator, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "separator"))
	if known && separator == "" {
		ctx.ReportNode(call, message)
	}
}

func (rule) Semantic() {}
