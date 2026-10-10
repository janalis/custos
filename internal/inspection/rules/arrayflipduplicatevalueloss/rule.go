// Package arrayflipduplicatevalueloss implements the native ArrayFlipDuplicateValueLoss inspection.
package arrayflipduplicatevalueloss

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve every key associated with this duplicate value."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayFlipDuplicateValueLoss" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_flip") {
		return
	}
	entries, ok := semanticquery.NativeArrayEntries(ctx, semanticquery.CallArgument(call.Args, 0, "array"))
	if !ok {
		return
	}
	seen := map[string]bool{}
	for _, value := range entries {
		key, known := semanticquery.NativeArrayKey(ctx, value)
		if !known {
			continue
		}
		if seen[key] {
			ctx.ReportNode(call, message)
			return
		}
		seen[key] = true
	}
}
