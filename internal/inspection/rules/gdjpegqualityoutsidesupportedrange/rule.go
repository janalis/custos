// Package gdjpegqualityoutsidesupportedrange implements the native GdJpegQualityOutsideSupportedRange inspection.
package gdjpegqualityoutsidesupportedrange

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply JPEG quality between zero and 100."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GdJpegQualityOutsideSupportedRange" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "imagejpeg") {
		return
	}
	quality := semanticquery.CallArgument(c.Args, 2, "quality")
	v, known := semanticquery.NativeContractInt(ctx, quality)
	if known && (v < -1 || v > 100) && !steppedBefore(ctx, quality) {
		ctx.ReportNode(n, message)
	}
}

// steppedBefore reports whether the variable passed as quality, or a variable
// it was copied from, can change ahead of the call without an assignment the
// value resolver models: `++`/`--`, a closure capturing it by reference, or
// extract()/parse_str() writing the local symbol table. The assigned literal
// is then no longer proof of the passed value.
func steppedBefore(ctx *analysis.Context, quality syntax.Expr) bool {
	variable, ok := syntax.UnwrapParens(quality).(*syntax.Variable)
	if !ok {
		return false
	}
	scope, limit := syntax.EnclosingVariableScope(quality), quality.Span().Start
	names := map[string]bool{variable.Name: true}
	found := false
	for grew := true; grew && !found; {
		grew = false
		visit := func(n syntax.Node) bool {
			if n.Span().Start >= limit {
				return false
			}
			switch n := n.(type) {
			case *syntax.IncDec:
				target, ok := syntax.UnwrapParens(n.Var).(*syntax.Variable)
				found = found || ok && names[target.Name]
			case *syntax.Assign:
				// The resolver follows `$quality = $base`, so $base matters too.
				target, tok := n.Var.(*syntax.Variable)
				source, sok := syntax.UnwrapParens(n.Value).(*syntax.Variable)
				if tok && sok && names[target.Name] && !names[source.Name] {
					names[source.Name], grew = true, true
				}
			case *syntax.Closure:
				for _, use := range n.Uses {
					found = found || use.ByRef && names[use.Var.Name]
				}
			case *syntax.FuncCall:
				name := semanticquery.NativeBuiltinName(ctx, n)
				found = found || name == "extract" || name == "parse_str" && semanticquery.CallArgument(n.Args, 1, "result") == nil
			}
			// Nested scopes own their locals.
			return !found && (n == scope || !syntax.IsVariableScope(n))
		}
		if scope == nil {
			syntax.InspectFile(ctx.File, visit)
		} else {
			syntax.Inspect(scope, visit)
		}
	}
	return found
}
