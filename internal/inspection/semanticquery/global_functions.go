package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// CallsGlobalFunction reports whether a named function call resolves to a
// function of the global namespace: fully qualified, unqualified in the
// global namespace, or unqualified in a namespace without a same-named
// namespaced function (PHP's runtime fallback). When known is set, the
// global function must also exist in the index.
func CallsGlobalFunction(ctx *analysis.Context, call *syntax.FuncCall, known bool) bool {
	name := ctx.GlobalFunctionName(call)
	return name != "" && (!known || ctx.Index().Function(name, ctx.PHP) != nil)
}

// NamesGlobalClass reports whether e is a class name that resolves, with the
// file's namespace and imports, to the global class fqn (any case, as PHP
// compares class names).
func NamesGlobalClass(ctx *analysis.Context, e syntax.Expr, fqn string) bool {
	n, ok := e.(*syntax.Name)
	if !ok {
		return false
	}
	return strings.EqualFold(ctx.Names().Class(n.Value, n.Span().Start), fqn)
}
