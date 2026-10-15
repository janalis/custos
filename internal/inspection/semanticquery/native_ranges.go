package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// NativeIntRange returns bounds including every reaching integer value.
// Exact native constants retain their existing resolution and version gates.
func NativeIntRange(ctx *analysis.Context, e syntax.Expr) (lo, hi int64, known bool) {
	if lo, hi, known = ctx.Flow().IntRange(e); known {
		return lo, hi, true
	}
	variable := false
	syntax.Inspect(e, func(n syntax.Node) bool {
		if _, ok := n.(*syntax.Variable); ok {
			variable = true
		}
		return !variable
	})
	if variable {
		return 0, 0, false
	}
	if n, ok := NativeContractInt(ctx, e); ok {
		return n, n, true
	}
	return 0, 0, false
}

// NativeIntEntry returns a concrete integer on a bounded CFG entry prefix.
// It does not treat widened bounds as witnesses of an actual execution.
func NativeIntEntry(ctx *analysis.Context, e syntax.Expr) (int64, bool) {
	return ctx.Flow().IntEntry(e)
}
