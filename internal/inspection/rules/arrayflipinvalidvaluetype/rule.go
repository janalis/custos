// Package arrayflipinvalidvaluetype implements the native ArrayFlipInvalidValueType inspection.
package arrayflipinvalidvaluetype

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply only string or integer values to array_flip."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayFlipInvalidValueType" }
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
	for _, value := range entries {
		t := ctx.TypeOf(value)
		if !t.IsUnknown() && !t.HasAny("mixed", "string", "int") {
			ctx.ReportNode(call, message)
			return
		}
	}
}
