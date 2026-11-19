// Package writetoreadonlystream implements the native WriteToReadOnlyStream inspection.
package writetoreadonlystream

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Open the stream with write access."

type rule struct{}

// New returns the WriteToReadOnlyStream inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "WriteToReadOnlyStream" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "fwrite") && !semanticquery.NativeBuiltin(ctx, c, "fputs") && !semanticquery.NativeBuiltin(ctx, c, "fputcsv") && !semanticquery.NativeBuiltin(ctx, c, "ftruncate") {
		return
	}
	handle := semanticquery.CallArgument(c.Args, 0, "stream")
	mode, known := semanticquery.NativeStreamMode(ctx, handle, c)
	if !known || mode[0] != 'r' || strings.Contains(mode, "+") {
		return
	}
	ctx.ReportNode(c, message)
}
