package infer

import "custos/internal/syntax"

// skippedExpressionAt identifies optional coalescing RHSs and side
// expressions of a nullsafe chain. Receiver evaluation itself always runs.
func (e *Env) skippedExpressionAt(parent syntax.Node, at uint32) bool {
	if a, ok := parent.(*syntax.Assign); ok && a.Op.Kind == syntax.TCoalesceEqual {
		return a.Value.Span().Start <= at && at < a.Value.Span().End
	}
	x, ok := parent.(syntax.Expr)
	if !ok {
		return false
	}
	receiver, safe := optionalChainParent(x)
	if receiver == nil || (receiver.Span().Start <= at && at < receiver.Span().End) {
		return false
	}
	if previous, _ := optionalChainParent(receiver); !safe && previous == nil {
		return false // ordinary calls and terminal receivers cannot skip
	}
	return e.hasNullsafeChain(x)
}

func optionalChainParent(x syntax.Expr) (syntax.Expr, bool) {
	if call, ok := x.(*syntax.FuncCall); ok {
		return call.Name, false
	}
	return chainParent(x)
}

// hasNullsafeChain caches syntactic receiver chains. Iterative collection
// avoids recursion and makes shared long chains linear across definitions.
func (e *Env) hasNullsafeChain(x syntax.Expr) bool {
	var path []syntax.Expr
	found := false
	for x != nil {
		if cached, ok := e.conditionalChains[x]; ok {
			found = cached
			break
		}
		path = append(path, x)
		receiver, safe := optionalChainParent(x)
		if safe {
			found = true
			break
		}
		x = receiver
	}
	if e.conditionalChains == nil {
		e.conditionalChains = map[syntax.Expr]bool{}
	}
	for _, expr := range path {
		e.conditionalChains[expr] = found
	}
	return found
}
