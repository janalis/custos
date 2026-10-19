package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// NativeNarrowedExcludes reports whether the conditions guarding read x prove
// it does not hold atom (`null`, `false`): its native type there, narrowed by
// those conditions, is known and lacks atom. A plain local is typed as read;
// a property, static property or literal-key element read narrows seed, the
// type its producer stores, since its declaration may say nothing. A static
// local is left to the flow analysis, which knows the calls that may enter
// the function again and assign it; other expressions prove nothing.
func NativeNarrowedExcludes(ctx *analysis.Context, x syntax.Expr, seed types.Type, atom string) bool {
	var t types.Type
	switch v := syntax.UnwrapParens(x).(type) {
	case *syntax.Variable:
		if v.Name == "" || v.NameExpr != nil || NativeStaticLocal(ctx, v) {
			return false
		}
		t = ctx.Types().Native().TypeOf(v)
	case *syntax.PropertyFetch, *syntax.StaticPropertyFetch, *syntax.ArrayDimFetch:
		t = ctx.Types().Native().NarrowMember(v, seed)
	default:
		return false
	}
	return !t.IsUnknown() && !t.Has(atom)
}
