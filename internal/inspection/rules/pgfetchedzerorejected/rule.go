// Package pgfetchedzerorejected implements PgFetchedZeroRejected.
package pgfetchedzerorejected

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Distinguish PostgreSQL fetch failure from a zero value."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PgFetchedZeroRejected" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pg_fetch_result") {
		return
	}
	var expr syntax.Expr = c
	if a, ok := c.Parent().(*syntax.Assign); ok && a.Value == c {
		expr = a
	}
	p := expr.Parent()
	switch p.(type) {
	case *syntax.If, *syntax.While:
	default:
		return
	}
	var fixes []diagnostic.Fix
	text := ctx.Text(expr)
	if !strings.Contains(text, "/*") && !strings.Contains(text, "//") && !strings.Contains(text, "#") {
		fixes = append(fixes, edit(expr.Span(), "("+text+")!==false"))
	}
	ctx.ReportNode(expr, message, fixes...)
}

func edit(span syntax.Span, text string) diagnostic.Fix {
	return diagnostic.Fix{Title: message, Edits: func() []diagnostic.TextEdit { return []diagnostic.TextEdit{{Span: span, NewText: text}} }}
}
