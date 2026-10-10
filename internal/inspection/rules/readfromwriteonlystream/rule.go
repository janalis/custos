// Package readfromwriteonlystream implements the native ReadFromWriteOnlyStream inspection.
package readfromwriteonlystream

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Open the stream with read access."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ReadFromWriteOnlyStream" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "fread") && !semanticquery.NativeBuiltin(ctx, c, "fgets") && !semanticquery.NativeBuiltin(ctx, c, "fgetc") && !semanticquery.NativeBuiltin(ctx, c, "fgetcsv") && !semanticquery.NativeBuiltin(ctx, c, "stream_get_contents") && !semanticquery.NativeBuiltin(ctx, c, "fpassthru") {
		return
	}
	mode, known := semanticquery.NativeStreamMode(ctx, semanticquery.CallArgument(c.Args, 0, "stream"), c)
	if known && mode != "" && mode[0] != 'r' && !strings.Contains(mode, "+") {
		ctx.ReportNode(c, message)
	}
}
