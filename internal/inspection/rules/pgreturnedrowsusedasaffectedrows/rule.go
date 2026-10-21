// Package pgreturnedrowsusedasaffectedrows implements PgReturnedRowsUsedAsAffectedRows.
package pgreturnedrowsusedasaffectedrows

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Use affected-row count for this statement."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PgReturnedRowsUsedAsAffectedRows" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pg_num_rows") {
		return
	}
	value := semanticquery.CallArgument(c.Args, 0, "result")
	query := semanticquery.OriginCall(ctx, value, "pg_query")
	if query == nil || !semanticquery.NativeFailureExcluded(ctx, value, types.Of(`\PgSql\Result`, "resource", "false"), "false") {
		return
	}
	sql, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(query.Args, 1, "query"))
	if !known {
		sql, known = semanticquery.NativeString(ctx, semanticquery.CallArgument(query.Args, 0, "query"))
	}
	if !known {
		return
	}
	tokens, known := semanticquery.SQLTokens(sql)
	if !known || len(tokens) == 0 {
		return
	}
	switch tokens[0] {
	case "INSERT", "UPDATE", "DELETE":
	default:
		return
	}
	for _, token := range tokens {
		if token == "RETURNING" || token == ";" {
			return
		}
	}
	ctx.ReportNode(c, message)
}
