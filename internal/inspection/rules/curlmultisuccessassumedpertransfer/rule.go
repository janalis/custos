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
	scope := syntax.EnclosingVariableScope(result)
	variable, ok := syntax.UnwrapParens(handle).(*syntax.Variable)
	if !ok || variable.Name == "" || handleEscapes(ctx, scope, variable.Name, branch.Span().Start) {
		return
	}
	for _, call := range ctx.Flow().Calls(scope) {
		if call.Node.Span().Start >= branch.Span().Start {
			break
		}
		fc, ok := call.Node.(*syntax.FuncCall)
		if ok && semanticquery.NativeBuiltin(ctx, fc, "curl_multi_add_handle") && semanticquery.NativeValue(ctx, semanticquery.CallArgument(fc.Args, 0, "multi_handle")) == origin && semanticquery.NativeDominates(fc, result) {
			ctx.ReportNode(result, message)
			return
		}
	}
}

// lifecycle lists the builtins that drive a multi handle without exposing
// per-transfer results and without detaching transfers from it.
var lifecycle = map[string]bool{
	"curl_multi_add_handle": true,
	"curl_multi_exec":       true,
	"curl_multi_select":     true,
	"curl_multi_setopt":     true,
	"curl_multi_errno":      true,
}

// handleEscapes reports whether the named handle variable is used before
// limit in any way other than being assigned or passed directly to a
// lifecycle builtin. Every other use (a helper or method receiving it, an
// alias, a capture, a stored copy, curl_multi_info_read itself) may inspect
// or change the transfers, so the local proof no longer holds.
func handleEscapes(ctx *analysis.Context, scope syntax.Node, name string, limit uint32) bool {
	escapes := false
	visit := func(n syntax.Node) bool {
		if escapes || n.Span().Start >= limit {
			return false
		}
		if call, ok := n.(*syntax.FuncCall); ok {
			// These builtins read the scope's variables by name.
			switch semanticquery.NativeBuiltinName(ctx, call) {
			case "compact", "get_defined_vars":
				escapes = true
			}
		}
		v, ok := n.(*syntax.Variable)
		if ok && v.NameExpr != nil {
			// A variable variable may name the handle.
			escapes = true
		}
		if !ok || v.Name != name {
			return true
		}
		switch parent := v.Parent().(type) {
		case *syntax.Assign:
			escapes = parent.Var != v || parent.ByRef
		case *syntax.Arg:
			call, ok := parent.Parent().Parent().(*syntax.FuncCall)
			escapes = !ok || parent.ByRef || parent.Unpack || !lifecycle[semanticquery.NativeBuiltinName(ctx, call)]
		default:
			escapes = true
		}
		return true
	}
	if scope == nil {
		// Top-level code has no enclosing scope node.
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	return escapes
}
