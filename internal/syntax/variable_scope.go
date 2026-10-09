package syntax

// IsVariableScope reports whether n owns local variables. Property hooks own
// locals independently of their class or promoting constructor.
func IsVariableScope(n Node) bool {
	if _, ok := n.(*PropertyHook); ok {
		return true
	}
	return IsFuncLike(n)
}

// EnclosingVariableScope returns the nearest strictly enclosing local scope.
func EnclosingVariableScope(n Node) Node {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if IsVariableScope(p) {
			return p
		}
	}
	return nil
}

// VariableScopeParams returns the explicit parameters of a local scope.
// The implicit setter parameter is not an AST parameter.
func VariableScopeParams(n Node) []*Param {
	if h, ok := n.(*PropertyHook); ok {
		return h.Params
	}
	return FuncLikeParams(n)
}

// VariableScopeBody returns a block or expression body, or nil for a scope
// without a body. Unlike FuncLikeBody it includes arrows and property hooks.
func VariableScopeBody(n Node) Node {
	switch n := n.(type) {
	case *PropertyHook:
		return n.Body
	case *ArrowFunction:
		return n.Expr
	default:
		if b := FuncLikeBody(n); b != nil {
			return b
		}
		return nil
	}
}
