// Package csvdelimiterinvalidbytelength implements the native CsvDelimiterInvalidByteLength inspection.
package csvdelimiterinvalidbytelength

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use a one-byte CSV separator."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CsvDelimiterInvalidByteLength" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall} }

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	args, pos, name, ok := semanticquery.NativeCSVSignature(ctx, n)
	if !ok {
		return
	}
	value, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(args, pos, name))
	if known && len(value) != 1 {
		ctx.ReportNode(n, message)
	}
}
