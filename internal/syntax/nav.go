package syntax

import "strings"

// Shared AST navigation helpers. They live here (not in analysis/util) so
// that the inference engine, which util depends on, uses the same code.

// UnwrapParens strips any number of enclosing parentheses from e (a Paren
// without an inner expression is returned as is).
func UnwrapParens(e Expr) Expr {
	for {
		p, ok := e.(*Paren)
		if !ok || p.Expr == nil {
			return e
		}
		e = p.Expr
	}
}

// IsFuncLike reports whether n is a function, method, closure or arrow
// function.
func IsFuncLike(n Node) bool {
	switch n.(type) {
	case *Function, *Method, *Closure, *ArrowFunction:
		return true
	}
	return false
}

// EnclosingFuncLike returns the nearest function, method, closure or arrow
// function strictly enclosing n (nil in top-level code).
func EnclosingFuncLike(n Node) Node {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if IsFuncLike(p) {
			return p
		}
	}
	return nil
}

// FuncLikeBody returns the braced body of a function-like (nil for arrow
// functions and abstract methods).
func FuncLikeBody(n Node) *Block {
	switch f := n.(type) {
	case *Function:
		return f.Body
	case *Method:
		return f.Body
	case *Closure:
		return f.Body
	}
	return nil
}

// FuncLikeParams returns the parameters of a function-like.
func FuncLikeParams(n Node) []*Param {
	switch f := n.(type) {
	case *Function:
		return f.Params
	case *Method:
		return f.Params
	case *Closure:
		return f.Params
	case *ArrowFunction:
		return f.Params
	}
	return nil
}

// EnclosingClass returns the class-like declaration containing n (nil
// outside one).
func EnclosingClass(n Node) *ClassLike {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if c, ok := p.(*ClassLike); ok {
			return c
		}
	}
	return nil
}

// IsNullConst reports whether e is the constant null (any case, optionally
// `\`-qualified). Parentheses are not looked through.
func IsNullConst(e Expr) bool {
	c, ok := e.(*ConstFetch)
	return ok && c.Name != nil && strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "null")
}
