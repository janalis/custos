// Package pgasyncresultsnotdrained implements PgAsyncResultsNotDrained.
package pgasyncresultsnotdrained

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Drain all pending PostgreSQL results before another asynchronous query."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PgAsyncResultsNotDrained" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pg_send_query") && !semanticquery.NativeBuiltin(ctx, c, "pg_send_query_params") {
		return
	}
	connection := semanticquery.CallArgument(c.Args, 0, "connection")
	if connection == nil {
		return
	}
	for p := c.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		branch, ok := p.(*syntax.If)
		if !ok || !branch.Body.Span().Contains(c.Span()) {
			continue
		}
		send, ok := syntax.UnwrapParens(branch.Cond).(*syntax.FuncCall)
		if !ok || !semanticquery.NativeBuiltin(ctx, send, "pg_send_query") || !same(ctx, connection, semanticquery.CallArgument(send.Args, 0, "connection")) {
			continue
		}
		sql, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(send.Args, 1, "query"))
		if !known {
			return
		}
		count := statementCount(sql)
		if count == 0 {
			return
		}
		body, ok := branch.Body.(*syntax.Block)
		if !ok {
			return
		}
		for _, s := range body.Stmts {
			if s.Span().Start >= c.Span().Start {
				break
			}
			st, ok := s.(*syntax.ExprStmt)
			if !ok {
				return
			}
			call, ok := st.Expr.(*syntax.FuncCall)
			if !ok || !semanticquery.NativeBuiltin(ctx, call, "pg_get_result") || !same(ctx, connection, semanticquery.CallArgument(call.Args, 0, "connection")) {
				return
			}
			count--
		}
		if count > 0 {
			ctx.ReportNode(c, message)
		}
		return
	}
}

func same(ctx *analysis.Context, a, b syntax.Expr) bool {
	return semanticquery.ExpansionSameObject(ctx, a, b)
}

func statementCount(sql string) int {
	if strings.ContainsAny(sql, `'\"$`) || strings.Contains(sql, "--") || strings.Contains(sql, "/*") {
		return 0
	}
	count := 0
	for _, s := range strings.Split(sql, ";") {
		if strings.TrimSpace(s) != "" {
			count++
		}
	}
	return count
}
