// Package pgescapedliteralquotedagain implements PgEscapedLiteralQuotedAgain.
package pgescapedliteralquotedagain

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use the already quoted PostgreSQL literal directly."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PgEscapedLiteralQuotedAgain" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pg_query") {
		return
	}
	query := semanticquery.CallArgument(c.Args, 1, "query")
	if query == nil {
		query = semanticquery.CallArgument(c.Args, 0, "query")
	}
	parts, values, known := semanticquery.ExpansionSQLParts(ctx, query)
	if !known {
		return
	}
	for i, v := range values {
		escape, ok := semanticquery.NativeValue(ctx, v).(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, escape, "pg_escape_literal") {
			continue
		}
		before := parts[i]
		after := parts[i+1]
		if strings.HasSuffix(before, "'") && strings.HasPrefix(after, "'") && quoteBoundary(before[:len(before)-1]) {
			ctx.ReportNode(query, message)
			return
		}
	}
}

func quoteBoundary(s string) bool {
	if strings.ContainsAny(s, `\";$`) || strings.Contains(s, "--") || strings.Contains(s, "/*") {
		return false
	}
	return strings.Count(s, "'")%2 == 0
}
