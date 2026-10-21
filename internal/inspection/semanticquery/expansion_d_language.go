package semanticquery

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// FiberConstruction returns the `new Fiber` that the receiver of c, a call
// of the Fiber method name, holds; nil before PHP 8.1 or when unknown.
func FiberConstruction(ctx *analysis.Context, c *syntax.MethodCall, name string) *syntax.New {
	if ctx.PHP < phpversion.PHP81 || !NativeMethod(ctx, c, "Fiber", name) {
		return nil
	}
	return NativeConstruction(ctx, c.Var, "Fiber")
}

// IsLiteral reports whether e is a literal or true, false or null.
func IsLiteral(e syntax.Expr) bool {
	switch value := syntax.UnwrapParens(e).(type) {
	case *syntax.Literal:
		return true
	case *syntax.ConstFetch:
		return value.Name != nil && (strings.EqualFold(value.Name.Value, "true") || strings.EqualFold(value.Name.Value, "false") || strings.EqualFold(value.Name.Value, "null"))
	}
	return false
}
