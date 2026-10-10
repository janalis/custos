// Package processexecfailurereportedassuccess implements the native ProcessExecFailureReportedAsSuccess inspection.
package processexecfailurereportedassuccess

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Report failure when process replacement returns."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ProcessExecFailureReportedAsSuccess" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "pcntl_exec") {
		return
	}
	st, ok := c.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	next, exists := astquery.NextStmt(ctx.File, st)
	if !exists {
		return
	}
	es, ok := next.(*syntax.ExprStmt)
	if !ok {
		return
	}
	e, ok := es.Expr.(*syntax.Exit)
	if !ok {
		return
	}
	a := semanticquery.CallArgument(e.Args, 0, "status")
	if a == nil {
		ctx.ReportNode(n, message)
		return
	}
	v, known := semanticquery.NativeContractInt(ctx, a)
	if known && v == 0 {
		ctx.ReportNode(n, message)
	}
}
