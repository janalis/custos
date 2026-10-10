// Package arraywalkcallbackreturnignored implements the native ArrayWalkCallbackReturnIgnored inspection.
package arraywalkcallbackreturnignored

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

const message = "Use array_map to retain the callback results."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayWalkCallbackReturnIgnored" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_walk") {
		return
	}
	callback := semanticquery.CallArgument(call.Args, 1, "callback")
	params, body := semanticquery.NativeCallback(ctx, callback)
	if len(params) == 0 || params[0].ByRef {
		return
	}
	var value syntax.Expr
	switch b := body.(type) {
	case syntax.Expr:
		value = b
	case *syntax.Block:
		if len(b.Stmts) != 1 {
			return
		}
		ret, ok := b.Stmts[0].(*syntax.Return)
		if !ok {
			return
		}
		value = ret.Expr
	}
	if pure(ctx, value, 0) {
		ctx.ReportNode(callback, message)
	}
}

func pure(ctx *analysis.Context, e syntax.Expr, depth int) bool {
	if e == nil || depth > 16 {
		return false
	}
	e = syntax.UnwrapParens(e)
	switch x := e.(type) {
	case *syntax.Variable:
		return x.Name != ""
	case *syntax.Literal:
		return true
	case *syntax.FuncCall:
		switch semanticquery.NativeBuiltinName(ctx, x) {
		case "strtoupper", "strtolower", "trim", "ltrim", "rtrim":
		default:
			return false
		}
		for _, node := range x.Args.Args {
			a, ok := node.(*syntax.Arg)
			if !ok || a.Unpack || !pure(ctx, a.Value, depth+1) || !ctx.TypeOf(a.Value).Equal(types.String) {
				return false
			}
		}
		return true
	}
	return false
}
