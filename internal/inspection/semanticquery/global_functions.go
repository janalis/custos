package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

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
