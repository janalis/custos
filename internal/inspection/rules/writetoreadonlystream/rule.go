// Package writetoreadonlystream implements the native WriteToReadOnlyStream inspection.
package writetoreadonlystream

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Open the stream with write access."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "WriteToReadOnlyStream" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "fwrite") && !semanticquery.NativeBuiltin(ctx, c, "fputs") && !semanticquery.NativeBuiltin(ctx, c, "fputcsv") && !semanticquery.NativeBuiltin(ctx, c, "ftruncate") {
		return
	}
	mode, known := semanticquery.NativeStreamMode(ctx, semanticquery.CallArgument(c.Args, 0, "stream"), c)
	if known && mode != "" && mode[0] == 'r' && !strings.Contains(mode, "+") {
		ctx.ReportNode(c, message)
	}
}
