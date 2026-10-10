// Package curlmultisuccessassumedpertransfer implements the native CurlMultiSuccessAssumedPerTransfer inspection.
package curlmultisuccessassumedpertransfer

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Inspect each completed transfer result."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlMultiSuccessAssumedPerTransfer" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KIf} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	branch := n.(*syntax.If)
	condition, ok := syntax.UnwrapParens(branch.Cond).(*syntax.Binary)
	if !ok || condition.Op.Kind != syntax.TBooleanAnd {
		return
	}
	check, ok := syntax.UnwrapParens(condition.Left).(*syntax.Binary)
	if !ok || check.Op.Kind != syntax.TIsIdentical && check.Op.Kind != syntax.TIsEqual {
		return
	}
	result, ok := semanticquery.NativeValue(ctx, check.Left).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, result, "curl_multi_exec") {
		return
	}
	constant, ok := check.Right.(*syntax.ConstFetch)
	if !ok || semanticquery.GlobalConstName(ctx, constant) != "CURLM_OK" {
		return
	}
	done, ok := syntax.UnwrapParens(condition.Right).(*syntax.Unary)
	if !ok || done.Op.Kind != syntax.TExclaim {
		return
	}
	running, ok := done.Expr.(*syntax.Variable)
	argument, ak := semanticquery.CallArgument(result.Args, 1, "still_running").(*syntax.Variable)
	if !ok || !ak || running.Name != argument.Name {
		return
	}
	block, ok := branch.Body.(*syntax.Block)
	if !ok || len(block.Stmts) != 1 {
		return
	}
	ret, ok := block.Stmts[0].(*syntax.Return)
	if !ok {
		return
	}
	truth, known := semanticquery.NativeTruth(ctx, ret.Expr)
	if !known || !truth {
		return
	}
	handle := semanticquery.CallArgument(result.Args, 0, "multi_handle")
	origin, ok := semanticquery.NativeValue(ctx, handle).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, origin, "curl_multi_init") {
		return
	}
	hasTransfer := false
	for _, call := range ctx.Flow().Calls(syntax.EnclosingVariableScope(result)) {
		if call.Node.Span().Start >= branch.Span().Start {
			break
		}
		fc, ok := call.Node.(*syntax.FuncCall)
		if !ok {
			continue
		}
		if semanticquery.NativeBuiltin(ctx, fc, "curl_multi_add_handle") && semanticquery.NativeValue(ctx, semanticquery.CallArgument(fc.Args, 0, "multi_handle")) == origin && semanticquery.NativeDominates(fc, result) {
			hasTransfer = true
		}
		if (semanticquery.NativeBuiltin(ctx, fc, "curl_multi_info_read") || semanticquery.NativeBuiltin(ctx, fc, "curl_multi_remove_handle")) && semanticquery.NativeValue(ctx, semanticquery.CallArgument(fc.Args, 0, "multi_handle")) == origin {
			return
		}
	}
	if hasTransfer {
		ctx.ReportNode(result, message)
	}
}
