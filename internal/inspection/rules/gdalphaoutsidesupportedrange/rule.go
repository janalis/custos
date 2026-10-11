// Package gdalphaoutsidesupportedrange implements the native GdAlphaOutsideSupportedRange inspection.
package gdalphaoutsidesupportedrange

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply GD alpha between zero and 127."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GdAlphaOutsideSupportedRange" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "imagecolorallocatealpha") && !semanticquery.NativeBuiltin(ctx, c, "imagecolorclosestalpha") {
		return
	}
	alpha := semanticquery.CallArgument(c.Args, 4, "alpha")
	v, known := semanticquery.NativeContractInt(ctx, alpha)
	if known && (v < 0 || v > 127) && !adjustedBefore(ctx, c, alpha) {
		ctx.ReportNode(n, message)
	}
}

// adjustedBefore reports whether the alpha variable, or a variable it was
// copied from, is stepped with ++/-- or captured by reference ahead of the
// call, or whether extract() runs ahead of it. The resolved literal then no longer describes the passed value.
func adjustedBefore(ctx *analysis.Context, call *syntax.FuncCall, alpha syntax.Expr) bool {
	variable, ok := syntax.UnwrapParens(alpha).(*syntax.Variable)
	if !ok {
		return false
	}
	scope := syntax.EnclosingVariableScope(call)
	limit := call.Span().Start
	names := map[string]bool{variable.Name: true}
	adjusted := false
	for grew := true; grew && !adjusted; {
		grew = false
		visit := func(n syntax.Node) bool {
			if n.Span().Start >= limit {
				return false
			}
			switch n := n.(type) {
			case *syntax.IncDec:
				if v, ok := syntax.UnwrapParens(n.Var).(*syntax.Variable); ok && names[v.Name] {
					adjusted = true
				}
			case *syntax.Assign:
				target, tok := n.Var.(*syntax.Variable)
				source, sok := syntax.UnwrapParens(n.Value).(*syntax.Variable)
				if tok && sok && names[target.Name] && !names[source.Name] {
					names[source.Name] = true
					grew = true
				}
			case *syntax.Closure:
				for _, use := range n.Uses {
					if use.ByRef && names[use.Var.Name] {
						adjusted = true
					}
				}
			case *syntax.FuncCall:
				// extract() can bind any local of the scope.
				if semanticquery.NativeBuiltinName(ctx, n) == "extract" {
					adjusted = true
				}
			}
			return n == scope || !syntax.IsVariableScope(n)
		}
		if scope == nil {
			syntax.InspectFile(ctx.File, visit)
		} else {
			syntax.Inspect(scope, visit)
		}
	}
	return adjusted
}
