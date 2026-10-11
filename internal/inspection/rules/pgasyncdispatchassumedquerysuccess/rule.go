// Package pgasyncdispatchassumedquerysuccess implements PgAsyncDispatchAssumedQuerySuccess.
package pgasyncdispatchassumedquerysuccess

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Inspect the asynchronous PostgreSQL result status."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PgAsyncDispatchAssumedQuerySuccess" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pg_send_query") && !semanticquery.NativeBuiltin(ctx, c, "pg_send_query_params") {
		return
	}
	branch, ok := c.Parent().(*syntax.If)
	if !ok || branch.Cond != c {
		return
	}
	// The guarded statement is the first one of a braced or alternative-syntax
	// body, or the body itself when the branch has no block.
	first := branch.Body
	if body, ok := first.(*syntax.Block); ok {
		if len(body.Stmts) == 0 {
			return
		}
		first = body.Stmts[0]
	}
	switch st := first.(type) {
	case *syntax.Return:
		truth, known := semanticquery.NativeTruth(ctx, st.Expr)
		if known && truth {
			ctx.ReportNode(c, message)
		}
	case *syntax.Echo:
		if len(st.Exprs) == 1 {
			if _, known := semanticquery.NativeString(ctx, st.Exprs[0]); known {
				ctx.ReportNode(c, message)
			}
		}
	}
}
