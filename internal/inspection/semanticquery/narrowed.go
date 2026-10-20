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
// a property, static property or element read (named by a literal or a
// local key) narrows seed, the
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
		t = ctx.Types().Native().NarrowedMember(v, seed)
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
	if !ok || NativeStaticLocal(ctx, v) || NativeLabelled(ctx, at) {
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
// `false`, `true`): t names it, or a type that includes it (`bool`, `mixed`).
func nativeMayHold(t types.Type, atom string) bool {
	return t.HasAny(atom, "mixed") || (atom == "false" || atom == "true") && t.Has("bool")
}

// NativeEarlierCallProves reports whether a call or comparison made earlier
// in the scope of at has a result that, narrowed by the conditions guarding
// at, satisfies the proves test pick returned for it. pick returns the type
// assumed for a node's result and that test, or a nil test to skip the node.
// An unknown narrowed result (at is unreachable, or the analysis gave up)
// also counts: callers use this to withhold a finding, never to raise one.
// Nested scopes are not searched.
func NativeEarlierCallProves(ctx *analysis.Context, at syntax.Expr, pick func(syntax.Expr) (seed types.Type, proves func(types.Type) bool)) bool {
	return NativeEarlierCallProvesIn(ctx, nil, at, pick)
}

// NativeEarlierCallProvesIn is NativeEarlierCallProves searching only root
// (a loop or branch body around at; nil for the whole scope).
func NativeEarlierCallProvesIn(ctx *analysis.Context, root syntax.Node, at syntax.Expr, pick func(syntax.Expr) (seed types.Type, proves func(types.Type) bool)) bool {
	env := ctx.Types().Native()
	found := false
	visit := func(n syntax.Node) bool {
		if found || n.Span().Start >= at.Span().Start || syntax.IsVariableScope(n) {
			return false
		}
		switch n.(type) {
		case *syntax.FuncCall, *syntax.MethodCall, *syntax.Binary:
			if n.Span().End > at.Span().Start {
				return true // a node holding at: its result comes later
			}
			if seed, proves := pick(n.(syntax.Expr)); proves != nil {
				t := env.CallResultAsAt(n.(syntax.Expr), at, seed)
				found = t.IsUnknown() || proves(t)
			}
		}
		return true
	}
	scope := syntax.EnclosingVariableScope(at)
	switch {
	case root != nil:
		syntax.Inspect(root, visit)
	case scope != nil:
		syntax.Inspect(syntax.VariableScopeBody(scope), visit)
	default:
		syntax.InspectFile(ctx.File, visit)
	}
	return found
}
