package syntax

import "strings"

// Shared AST navigation helpers. They live here (not in inspection queries) so
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

// IsFuncLike reports whether n is a function, method, closure, arrow
// function or property hook.
func IsFuncLike(n Node) bool {
	switch n.(type) {
	case *Function, *Method, *Closure, *ArrowFunction, *PropertyHook:
		return true
	}
	return false
}

// EnclosingFuncLike returns the nearest function-like strictly enclosing n (nil in top-level code).
func EnclosingFuncLike(n Node) Node {
	for p := n.Parent(); p != nil; p = p.Parent() {
		if IsFuncLike(p) {
			return p
		}
	}
	return nil
}

// FuncLikeBody returns the braced body of a function-like (nil for arrow
// functions, expression hooks and abstract declarations).
func FuncLikeBody(n Node) *Block {
	switch f := n.(type) {
	case *Function:
		return f.Body
	case *Method:
		return f.Body
	case *Closure:
		return f.Body
	case *PropertyHook:
		b, _ := f.Body.(*Block)
		return b
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
	case *PropertyHook:
		return f.Params
	}
	return nil
}

// EnclosingClass returns the class-like declaration containing n (nil
// outside one).
func EnclosingClass(n Node) *ClassLike {
	child := n
	for p := n.Parent(); p != nil; child, p = p, p.Parent() {
		// The constructor arguments of an anonymous class are evaluated
		// where the class is created, not inside it.
		if c, ok := p.(*ClassLike); ok && (c.Args == nil || child != Node(c.Args)) {
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

// Partial reports whether a holds a placeholder: the call only creates a
// closure and invokes nothing. That covers first-class callables
// (`f(...)`) and partial function applications (`f(?, 2)`, `f(1, ...)`,
// PHP 8.6).
func (a *ArgList) Partial() bool {
	if a == nil {
		return false
	}
	for _, x := range a.Args {
		switch x.(type) {
		case *VariadicPlaceholder, *ArgPlaceholder:
			return true
		}
	}
	return false
}

// FirstClassCallable reports whether a is the `(...)` of a first-class
// callable (`f(...)`, `$o->m(...)`, `C::m(...)`): the closure keeps the
// callee's whole signature. See Partial for every closure-creating list.
func (a *ArgList) FirstClassCallable() bool {
	if a == nil || len(a.Args) != 1 {
		return false
	}
	_, ok := a.Args[0].(*VariadicPlaceholder)
	return ok
}

// RecoveryError returns the first error the parser recovered from, which
// may have dropped or reshaped code; compile errors leave the tree whole.
func (f *File) RecoveryError() (Error, bool) {
	for _, e := range f.Errors {
		if !e.Compile {
			return e, true
		}
	}
	return Error{}, false
}

// Recovered reports whether the parser recovered from an error in f: its
// tree may lack or misplace code. A file with only compile errors is whole.
func (f *File) Recovered() bool {
	_, ok := f.RecoveryError()
	return ok
}

// RecoveryErrorCount is the number of errors the parser recovered from,
// compile errors aside.
func (f *File) RecoveryErrorCount() int {
	n := 0
	for _, e := range f.Errors {
		if !e.Compile {
			n++
		}
	}
	return n
}

// ParsesWorse reports whether f, a changed version of orig, parses worse:
// more errors, or more the parser recovered from, as when a parse error
// takes the place of a compile error at the same position.
func (f *File) ParsesWorse(orig *File) bool {
	return len(f.Errors) > len(orig.Errors) || f.RecoveryErrorCount() > orig.RecoveryErrorCount()
}
