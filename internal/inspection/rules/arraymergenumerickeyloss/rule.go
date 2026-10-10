// Package arraymergenumerickeyloss implements the native ArrayMergeNumericKeyLoss inspection.
package arraymergenumerickeyloss

import (
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve numeric identifier keys when combining arrays."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayMergeNumericKeyLoss" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_merge") {
		return
	}
	args, ok := astquery.CallArgValues(call)
	if !ok {
		return
	}
	for _, arg := range args {
		a := semanticquery.NativeArray(ctx, arg)
		if a == nil {
			continue
		}
		for i, item := range a.Items {
			if item.Key == nil {
				continue
			}
			key, known := semanticquery.NativeArrayKey(ctx, item.Key)
			if known && strings.HasPrefix(key, "i:") && key != "i:"+strconv.Itoa(i) {
				ctx.ReportNode(call, message)
				return
			}
		}
	}
}
