package semanticquery

import (
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// NativeNarrowedExcludes reports whether the conditions guarding read x prove
// it does not hold atom (`null`, `false`): its native type there, narrowed by
// those conditions, is known and lacks atom. Nothing is proven where a goto
// label lets a jump land between a guard and the read. A plain local is typed
// as read (narrowing seed when that type is unknown);
// a property, static property or literal-key element read narrows seed, the
// type its producer stores, since its declaration may say nothing. A static
// local is left to the flow analysis, which knows the calls that may enter
// the function again and assign it; other expressions prove nothing.
func NativeNarrowedExcludes(ctx *analysis.Context, x syntax.Expr, seed types.Type, atom string) bool {
	if NativeLabelled(ctx, x) {
		return false // a jump may land between a guard and the read
	}
	var t types.Type
	switch v := syntax.UnwrapParens(x).(type) {
	case *syntax.Variable:
		if v.Name == "" || v.NameExpr != nil || NativeStaticLocal(ctx, v) {
			return false
		}
		if t = ctx.Types().Native().TypeOf(v); t.IsUnknown() {
			// Merged with a value of unknown type: what the guards prove
			// holds whatever reaches the read.
			t = ctx.Types().Native().NarrowedFrom(v, seed)
		}
	case *syntax.PropertyFetch, *syntax.StaticPropertyFetch, *syntax.ArrayDimFetch:
		t = ctx.Types().Native().NarrowMember(v, seed)
	default:
		return false
	}
	return !t.IsUnknown() && !nativeMayHold(t, atom)
}

// NativeNarrowedExcludesAt reports whether the conditions guarding at, a
// later expression that read x still reaches with the same value (a call
// consuming what was computed from x), prove x does not hold atom there. Only
// a plain local that is not static is followed.
func NativeNarrowedExcludesAt(ctx *analysis.Context, x, at syntax.Expr, atom string) bool {
	v, ok := syntax.UnwrapParens(x).(*syntax.Variable)
	if !ok || NativeStaticLocal(ctx, v) {
		return false
	}
	env := ctx.Types().Native()
	if !env.ReachesUnchanged(v, at) {
		return false
	}
	t := env.TypeOfVarAt(v.Name, at)
	return !t.IsUnknown() && !nativeMayHold(t, atom)
}

// nativeMayHold reports whether a value of type t may be atom (`null`,
// `false`): t names it, or a type that includes it (`bool`, `mixed`).
func nativeMayHold(t types.Type, atom string) bool {
	return t.HasAny(atom, "mixed") || atom == "false" && t.Has("bool")
}
