package infer

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

// ForeachValueType is the type of the value loop fe binds on each
// iteration, as a local foreach target is typed; unknown when it cannot be
// told.
func (e *Env) ForeachValueType(fe *syntax.Foreach) types.Type { return e.foreachElem(fe.Expr) }

// ForeachKeyType is the type of the key loop fe binds on each iteration:
// that of a local key target for an array, the bound key type of a
// Traversable, else unknown (an iterator may yield keys of any type).
func (e *Env) ForeachKeyType(fe *syntax.Foreach) types.Type {
	t := e.baseType(fe.Expr)
	if t.IsUnknown() {
		return types.Unknown
	}
	if !t.Without("null", "false").IsArrayLike() {
		k, _ := e.iterTypes(t)
		return k
	}
	return e.foreachKey(fe.Expr)
}

// IncDecStoredType is the type of the value ++/-- n stores, for its prefix
// and postfix forms alike.
func (e *Env) IncDecStoredType(n *syntax.IncDec) types.Type { return e.incDecStored(n) }
