// Package processexitcodelostbyrepeatedstatusread implements the native ProcessExitCodeLostByRepeatedStatusRead inspection.
package processexitcodelostbyrepeatedstatusread

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Keep the exit code from the first completed status."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ProcessExitCodeLostByRepeatedStatusRead" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	if ctx.PHP >= phpversion.PHP83 {
		return
	}
	a := n.(*syntax.ArrayDimFetch)
	key, known := semanticquery.NativeString(ctx, a.Dim)
	if !known || key != "exitcode" {
		return
	}
	second, ok := semanticquery.NativeLocalValue(ctx, a.Var).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, second, "proc_get_status") {
		return
	}
	if !semanticquery.ExpansionDUnaliased(ctx, semanticquery.CallArgument(second.Args, 0, "process"), n) {
		return
	}
	for p := n.Parent(); p != nil && !syntax.IsVariableScope(p); p = p.Parent() {
		b, ok := p.(*syntax.If)
		if !ok || !b.Body.Span().Contains(n.Span()) {
			continue
		}
		u, ok := syntax.UnwrapParens(b.Cond).(*syntax.Unary)
		if !ok || u.Op.Kind != syntax.TExclaim {
			continue
		}
		running, ok := syntax.UnwrapParens(u.Expr).(*syntax.ArrayDimFetch)
		if !ok {
			continue
		}
		key, k := semanticquery.NativeString(ctx, running.Dim)
		if !k || key != "running" {
			continue
		}
		first, ok := semanticquery.NativeLocalValue(ctx, running.Var).(*syntax.FuncCall)
		if ok && first != second && semanticquery.NativeBuiltin(ctx, first, "proc_get_status") && semanticquery.NativeSameValue(ctx, semanticquery.CallArgument(first.Args, 0, "process"), semanticquery.CallArgument(second.Args, 0, "process")) {
			ctx.ReportNode(n, message)
			return
		}
	}
}
