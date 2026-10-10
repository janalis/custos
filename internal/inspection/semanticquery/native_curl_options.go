package semanticquery

import (
	"strconv"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
)

// NativeCurlOption returns a proven, explicitly configured cURL option value.
// Unknown updates, conditional setters and escapes invalidate the proof.
func NativeCurlOption(ctx *analysis.Context, at syntax.Node, handle syntax.Expr, option string) (syntax.Expr, bool) {
	value := ctx.Flow().Value(handle)
	if !value.Complete || value.Identity == 0 {
		return nil, false
	}
	origin, known := ctx.Flow().Resolve(handle)
	init, ok := origin.(*syntax.FuncCall)
	if !known || !ok || !NativeBuiltin(ctx, init, "curl_init") {
		return nil, false
	}
	scope := syntax.EnclosingVariableScope(at)
	key := "curl.calls:global"
	if scope != nil {
		key = "curl.calls:" + strconv.FormatUint(uint64(scope.Span().Start), 10)
	}
	index := ctx.Memo(key, func() any {
		out := map[uint32][]flow.Call{}
		for _, call := range ctx.Flow().Calls(scope) {
			seen := map[uint32]bool{}
			for _, arg := range call.Arguments {
				if arg.Identity != 0 && !seen[arg.Identity] {
					out[arg.Identity] = append(out[arg.Identity], call)
					seen[arg.Identity] = true
				}
			}
		}
		return out
	}).(map[uint32][]flow.Call)
	var result syntax.Expr
	for _, call := range index[value.Identity] {
		if call.Node.Span().Start >= at.Span().Start {
			break
		}
		fc, ok := call.Node.(*syntax.FuncCall)
		if !ok || !NativeBuiltin(ctx, fc, "curl_setopt") || !NativeDominates(fc, at) {
			return nil, false
		}
		key, ok := syntax.UnwrapParens(CallArgument(fc.Args, 1, "option")).(*syntax.ConstFetch)
		if !ok {
			return nil, false
		}
		name := GlobalConstName(ctx, key)
		constant := ctx.Index().Constant(name, ctx.PHP)
		if constant == nil || !constant.Builtin {
			return nil, false
		}
		if name == option {
			result = CallArgument(fc.Args, 2, "value")
			if result == nil {
				return nil, false
			}
		}
	}
	return result, result != nil
}
