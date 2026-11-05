package infer

import "custos/internal/php/syntax"

// Follow one unaliased reaching assignment at the read point. Declaration
// annotations, reference bindings, element mutations and loop-carried producers
// cannot establish the runtime entries of an argument pack.
func (e *Env) localArgumentPack(expr syntax.Expr, depth int) (*syntax.Array, bool) {
	if depth == 0 {
		return nil, false
	}
	switch n := syntax.UnwrapParens(expr).(type) {
	case *syntax.Array:
		return n, true
	case *syntax.Variable:
		if n.Name == "" || n.NameExpr != nil || n.Name == "this" {
			return nil, false
		}
		scope := syntax.EnclosingVariableScope(n)
		vars := e.scopeVars(scope)
		if vars.shared[n.Name] || vars.refs[n.Name] || len(vars.defs[n.Name]) > maxVarDefs {
			return nil, false
		}
		fwd, back, _ := e.reaching(vars.defs[n.Name], n, scope)
		fwd = e.changing(fwd)
		if len(fwd) != 1 || len(back) != 0 || fwd[0].asg == nil || fwd[0].asg.ByRef || fwd[0].asg.Op.Kind != syntax.TEqual {
			return nil, false
		}
		if writes, _ := e.reachingWritesAt(n.Name, n, scope); len(writes) != 0 {
			return nil, false
		}
		from := max(fwd[0].pos, fwd[0].end)
		if e.clobbered(vars, fwd, back, n, scope) || e.maybeUndefinedAt(vars, vars.defs[n.Name], fwd, n.Name, n, scope) || e.shapeClobberedAt(scope, n.Name, from, n) || e.nonEmptyBroken(scope, n.Name, from, n) {
			return nil, false
		}
		return e.localArgumentPack(fwd[0].asg.Value, depth-1)
	}
	return nil, false
}
