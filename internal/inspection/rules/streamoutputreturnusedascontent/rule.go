// Package streamoutputreturnusedascontent implements the native StreamOutputReturnUsedAsContent inspection.
package streamoutputreturnusedascontent

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Output the stream without printing its byte count."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "StreamOutputReturnUsedAsContent" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "readfile") && !semanticquery.NativeBuiltin(ctx, c, "fpassthru") {
		return
	}
	var wrapper syntax.Node
	var stmt syntax.Node
	// Grouping parentheses around the call (`echo(readfile($f));`) do not change
	// what is printed: look through them and keep them when the keyword goes.
	parent, operand := astquery.ParentSkipParens(c)
	switch p := parent.(type) {
	case *syntax.Echo:
		if len(p.Exprs) != 1 || p.Short {
			return
		}
		wrapper = p
		stmt = p
	case *syntax.Print:
		st, ok := p.Parent().(*syntax.ExprStmt)
		if !ok {
			return
		}
		wrapper = p
		stmt = st
	default:
		return
	}
	span := syntax.Span{Start: wrapper.Span().Start, End: operand.Span().Start}
	prefix := ctx.SpanText(span)
	if strings.Contains(prefix, "/*") || strings.Contains(prefix, "//") || strings.Contains(prefix, "#") {
		ctx.ReportNode(c, message)
		return
	}
	fix := diagnostic.Fix{Title: "Output the stream directly", Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{{Span: span, NewText: ""}}
	}}
	ctx.ReportNode(stmt, message, fix)
}
