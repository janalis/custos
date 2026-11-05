package flow

import "custos/internal/php/syntax"

var nullScalar = constantScalar{k: 'n'}

func nullResult() Value { return Value{Complete: true, literal: &nullScalar, plain: true, deep: true} }

// Only reaching the lexical end adds an implicit null. Explicit returns share
// the normal exit for effects, but do not imply that the end is reachable.
func completedReturn(scope syntax.Node, r *scopeResult, value Value) Value {
	if _, arrow := scope.(*syntax.ArrowFunction); arrow {
		return value
	}
	if !r.complete || r.generator || r.leaves || syntax.VariableScopeBody(scope) == nil {
		return value
	}
	if len(r.returns) == 0 {
		return nullResult()
	}
	if r.fallsThrough {
		return merge(value, nullResult())
	}
	return value
}
