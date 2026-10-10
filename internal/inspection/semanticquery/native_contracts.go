package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
)

// NativeStringInputs returns arguments bound to a resolved nonnullable string
// parameter. Unpack and unresolved contracts cannot establish a string use.
func NativeStringInputs(ctx *analysis.Context, call syntax.Node, list *syntax.ArgList) []syntax.Expr {
	callee, ok := ResolveCallee(ctx, call)
	if !ok {
		return nil
	}
	var inputs []syntax.Expr
	for i, p := range callee.Params {
		if p.Type == "string" {
			if arg := CallArgument(list, i, p.Name); arg != nil {
				inputs = append(inputs, arg)
			}
		}
	}
	return inputs
}

// NativeSuccessFollower finds an immediate unconditional true return after a
// discarded operation. It deliberately does not infer promises from names.
func NativeSuccessFollower(ctx *analysis.Context, call syntax.Node) bool {
	st, ok := call.Parent().(*syntax.ExprStmt)
	if !ok {
		return false
	}
	next, ok := astquery.NextStmt(ctx.File, st)
	if !ok {
		return false
	}
	ret, ok := next.(*syntax.Return)
	if !ok {
		return false
	}
	truth, known := astquery.BoolConst(ret.Expr)
	return known && truth
}
