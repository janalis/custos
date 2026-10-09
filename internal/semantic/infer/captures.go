package infer

import (
	"custos/internal/php/syntax"
	"custos/internal/semantic/types"
)

type captureKey struct {
	scope syntax.Expr
	name  string
}

// captureType evaluates a by-value import at closure construction in its
// enclosing scope. The source AST and its caches are left untouched.
func (e *Env) captureType(closure syntax.Expr, name string) types.Type {
	k := captureKey{closure, name}
	if t, ok := e.captures[k]; ok {
		return t
	}
	if e.captureBusy[k] || e.captureDepth >= maxCaptureDepth {
		return types.Unknown
	}
	if e.captures == nil {
		e.captures = map[captureKey]types.Type{}
		e.captureBusy = map[captureKey]bool{}
	}
	e.captureBusy[k] = true
	e.captureDepth++
	t := e.captureSnapshot(closure, name)
	e.captureDepth--
	delete(e.captureBusy, k)
	e.captures[k] = t
	return t
}

const maxCaptureDepth = 64

func (e *Env) captureSnapshot(closure syntax.Expr, name string) types.Type {
	scope := syntax.EnclosingVariableScope(closure)
	sv := e.scopeVars(scope)
	defs := sv.defs[name]
	if len(defs) > maxVarDefs {
		return types.Unknown
	}
	if arrow, ok := scope.(*syntax.ArrowFunction); ok && len(defs) == 0 {
		return e.captureType(arrow, name)
	}
	fwd, back, from := e.reaching(defs, closure, scope)
	if e.clobbered(sv, fwd, back, closure, scope) {
		return types.Unknown
	}
	var ts []types.Type
	after := uint32(0)
	for _, d := range fwd {
		ts = append(ts, d.typ())
		after = max(after, d.pos, d.end)
	}
	for _, d := range back {
		ts = append(ts, d.typ())
	}
	if len(ts) == 0 {
		return types.Unknown
	}
	if e.maybeUndefinedAt(sv, defs, fwd, name, closure, scope) {
		ts = append(ts, types.Null)
	}
	t := types.Union(ts...)
	if t.HasShape() && (e.shapeClobbered(scope, name) || e.nonEmptyBroken(scope, name, from, closure)) {
		t = t.WithoutShape()
	}
	if !t.ArrayKey().IsUnknown() && e.shapeClobbered(scope, name) {
		t = t.WithArrayKey(types.Unknown)
	}
	if t.IsNonEmptyArray() && e.nonEmptyBroken(scope, name, from, closure) {
		t = t.WithNonEmpty(false)
	}
	t = e.narrowExprAfter(t, closure, name, scope, after)
	t, _ = e.withElemWritesAt(t, name, closure, scope)
	return t
}
