// Package arraychunkinvalidsize implements the native ArrayChunkInvalidSize inspection.
package arraychunkinvalidsize

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply a positive chunk size."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayChunkInvalidSize" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_chunk") {
		return
	}
	size, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 1, "length"))
	if known && size < 1 {
		ctx.ReportNode(call, message)
	}
}
