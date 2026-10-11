// Package pgfetchedzerorejected implements PgFetchedZeroRejected.
package pgfetchedzerorejected

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const (
	message = "Distinguish PostgreSQL fetch failure from a zero value."
	title   = "Test the fetched value with is_string()"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PgFetchedZeroRejected" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pg_fetch_result") || c.Args == nil || astquery.WrittenArguments(c.Args) == nil {
		return // not the builtin, or a first-class callable that fetches nothing
	}
	var expr syntax.Expr = c
	// Only a plain by-value assignment to a simple variable yields the fetched
	// value unchanged: compound operators, typed properties and destructuring
	// test something else.
	if a, ok := c.Parent().(*syntax.Assign); ok && a.Value == c && a.Op.Kind == syntax.TEqual && !a.ByRef {
		if v, ok := a.Var.(*syntax.Variable); ok && v.Name != "" {
			expr = a
		}
	}
	switch expr.Parent().(type) {
	case *syntax.If, *syntax.While:
	default:
		return
	}
	// is_string() admits "0" and "" while still rejecting both the false of a
	// failed fetch and the null of an SQL NULL, as the truthiness test did.
	span := expr.Span()
	open := semanticquery.QualifiedBuiltin(ctx, "is_string", span.Start) + "("
	ctx.ReportNode(expr, message, diagnostic.Fix{Title: title, Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{
			{Span: syntax.Span{Start: span.Start, End: span.Start}, NewText: open},
			{Span: syntax.Span{Start: span.End, End: span.End}, NewText: ")"},
		}
	}})
}
