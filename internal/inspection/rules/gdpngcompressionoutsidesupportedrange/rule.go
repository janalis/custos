// Package gdpngcompressionoutsidesupportedrange implements the native GdPngCompressionOutsideSupportedRange inspection.
package gdpngcompressionoutsidesupportedrange

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply PNG compression between zero and nine."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "GdPngCompressionOutsideSupportedRange" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "imagepng") {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 2, "quality")
	v, known := semanticquery.NativeContractInt(ctx, arg)
	if known && (v < -1 || v > 9) && !rewrittenOutOfBand(ctx, c, arg) {
		ctx.ReportNode(n, message)
	}
}

// rewrittenOutOfBand reports whether a variable of the level expression, or a
// variable it was assigned from, can change without a plain assignment in the
// scope of the call: through `++` or `--`, through a closure or array item
// holding a reference to it, or through extract(). The resolved literal is
// then not a proven level. Arguments of unresolved or by-reference calls are
// already refused by the value resolver.
func rewrittenOutOfBand(ctx *analysis.Context, call *syntax.FuncCall, level syntax.Expr) bool {
	scope := syntax.EnclosingFuncLike(call)
	seen := map[string]bool{}
	var names []string
	collect := func(e syntax.Node) {
		syntax.Inspect(e, func(x syntax.Node) bool {
			if v, ok := x.(*syntax.Variable); ok && v.NameExpr == nil && !seen[v.Name] {
				seen[v.Name] = true
				names = append(names, v.Name)
			}
			return true
		})
	}
	collect(level)
	for i := 0; i < len(names); i++ {
		for _, access := range flowquery.VarAccesses(ctx.File, scope, names[i]) {
			switch by := access.By.(type) {
			case *syntax.IncDec:
				return true
			case *syntax.ClosureUse:
				if by.ByRef {
					return true
				}
			case *syntax.Assign:
				// The resolver follows `$level = $base`, so $base matters too.
				collect(by.Value)
			}
			if item, ok := access.Var.Parent().(*syntax.ArrayItem); ok && item.ByRef {
				return true
			}
		}
	}
	return len(names) > 0 && astquery.CallsExtract(ctx.File, scope)
}
