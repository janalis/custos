package infer

import (
	"custos/internal/index"
	"custos/internal/syntax"
	"custos/internal/types"
)

// BodyReturnType infers the return type of a function declared in this file
// from its `return` statements (nested closures/classes excluded): the union
// of the returned expressions' types, plus null for a bare `return;` or when
// the body has no return. Unknown when f is not declared in this file, has a
// declared or documented return type, is a generator, or any returned
// expression's type is unknown. Opt-in: TypeOf never uses it implicitly.
func (e *Env) BodyReturnType(f *index.Function) types.Type {
	if f == nil || f.Return != "" || f.DocReturn != "" || f.File != e.File.Path {
		return types.Unknown
	}
	var decl *syntax.Function
	syntax.InspectFile(e.File, func(n syntax.Node) bool {
		if decl != nil {
			return false
		}
		if fn, ok := n.(*syntax.Function); ok && fn.Span() == f.Span {
			decl = fn
			return false
		}
		return true
	})
	if decl == nil || decl.Body == nil {
		return types.Unknown
	}
	var ts []types.Type
	ok, sawReturn := true, false
	syntax.Inspect(decl.Body, func(n syntax.Node) bool {
		if !ok {
			return false
		}
		switch n := n.(type) {
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.Yield, *syntax.YieldFrom:
			ok = false
		case *syntax.Return:
			sawReturn = true
			if n.Expr == nil {
				ts = append(ts, types.Null)
				return false
			}
			t := e.TypeOf(n.Expr)
			if t.IsUnknown() {
				ok = false
			}
			ts = append(ts, t)
		}
		return true
	})
	if !ok {
		return types.Unknown
	}
	if !sawReturn {
		return types.Null
	}
	return types.Union(ts...)
}

// LiteralTextType infers the type of a constant/default initialiser from its
// source text (literals, arrays, true/false/null); unknown otherwise.
func LiteralTextType(v string) types.Type { return literalTextType(v) }
