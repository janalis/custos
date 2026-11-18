// Package explodeemptyseparator implements the native ExplodeEmptySeparator
// inspection.
package explodeemptyseparator

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

// New returns the ExplodeEmptySeparator inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ExplodeEmptySeparator" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Supply a nonempty separator."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "explode")
	if call == nil {
		return
	}
	arg := semanticquery.CallArgument(call.Args, 0, "separator")
	separator, known := semanticquery.NativeString(ctx, arg)
	if known && separator == "" && semanticquery.StableLocalValue(ctx, arg, semanticquery.LocalStability{}) {
		ctx.ReportNode(call, message)
	}
}

func (rule) Semantic() {}
