// Package signalwaitfailureacceptedassignal implements the native SignalWaitFailureAcceptedAsSignal inspection.
package signalwaitfailureacceptedassignal

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reject signal-wait failure before handling the result."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "SignalWaitFailureAcceptedAsSignal" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	b := n.(*syntax.If)
	e := syntax.UnwrapParens(b.Cond)
	var assigned syntax.Expr
	if a, ok := e.(*syntax.Assign); ok {
		assigned = a.Var
		e = a.Value
	}
	c, ok := semanticquery.NativeLocalValue(ctx, e).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, c, "pcntl_sigwaitinfo") {
		return
	}
	calls := ctx.Flow().Calls(syntax.EnclosingVariableScope(n))
	if len(calls) > 256 {
		return
	}
	for _, record := range calls {
		call, ok := record.Node.(*syntax.FuncCall)
		if !ok || !b.Body.Span().Contains(call.Span()) || call.Args == nil {
			continue
		}
		for _, node := range call.Args.Args {
			a, ok := node.(*syntax.Arg)
			if ok && !a.Unpack && (semanticquery.NativeLocalValue(ctx, a.Value) == c || (assigned != nil && directFirst(b.Body, call) && astquery.Equivalent(ctx.File, assigned, a.Value))) {
				ctx.ReportNode(b.Cond, message)
				return
			}
		}
	}
}

func directFirst(body syntax.Stmt, call *syntax.FuncCall) bool {
	if block, ok := body.(*syntax.Block); ok {
		body = block.Stmts[0]
	}
	st, ok := body.(*syntax.ExprStmt)
	return ok && st.Expr == call
}
