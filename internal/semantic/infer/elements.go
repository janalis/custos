package infer

import (
	"slices"

	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// sealedComputedDim types an element read with a computed key on an empty
// literal array: one of the values written into it since (a missing key
// reads null with a warning, which is not added, as for `T[]` elements). Unknown with an unknown write; ok is false for
// another type, or nothing known.
func (e *Env) sealedComputedDim(ct types.Type, n *syntax.ArrayDimFetch) (types.Type, bool) {
	// (A literal key on a sealed shape is shapeDim's: dimType asks it
	// first.) Only an empty literal: with listed keys, any of their values
	// may be read, a union rules would distrust.
	if !ct.IsSealedShape() || len(ct.ShapeKeys()) > 0 || n.Dim == nil || !ct.Without("null").IsArrayLike() {
		return types.Unknown, false
	}
	el := e.widenElem(types.Of("never"), n) // the values written since
	if el.IsUnknown() || el.OnlyOf("never") {
		return types.Unknown, false
	}
	if v := asVariable(n.Var); v != nil {
		if ws, _ := e.reachingWrites(v); slices.ContainsFunc(ws, func(w *elemWrite) bool { return w.nested }) {
			// `$by[$k][] = $row`: the stored arrays grew since.
			el = el.WithoutArrayInfo()
		}
	}
	return el, true
}

// asVariable returns x as a plain variable (nil otherwise).
func asVariable(x syntax.Expr) *syntax.Variable {
	if v, ok := syntax.UnwrapParens(x).(*syntax.Variable); ok && v.Name != "" && v.Name != "this" {
		return v
	}
	return nil
}

// dimType is the type of element read n before narrowing.
func (e *Env) dimType(n *syntax.ArrayDimFetch) types.Type {
	ct := e.baseType(n.Var)
	if fact, ok := e.chains[n.Var]; ok {
		ct = fact.evaluated
	}
	if ct.IsUnknown() {
		return types.Unknown
	}
	if t, ok := e.shapeDim(ct, n); ok {
		return t
	}
	if t, ok := e.sealedComputedDim(ct, n); ok {
		return t
	}
	// X[]|null (or |false: a failed builtin) indexes to X; a plain `array`
	// member has unknown elements.
	if el := ct.Elem(); !el.IsUnknown() && ct.Without("null", "false").IsArrayLike() && !ct.Has("array") {
		return e.widenElem(el, n)
	}
	if ct.Without("null", "false").OnlyOf("string") {
		return types.String
	}
	return e.arrayAccessDimType(n, ct)
}
