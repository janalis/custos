package infer

import (
	"strings"

	"custos/internal/index"
	"custos/internal/syntax"
	"custos/internal/types"
)

// maxBodyDepth caps how many body inferences may be nested (a call in a
// returned expression whose callee's body is inferred, and so on).
const maxBodyDepth = 4

// BodyReturnType infers the return type of a function declared in this file
// from its body (see bodyReturn). Unknown when f is not declared in this
// file or has a declared or documented return type.
func (e *Env) BodyReturnType(f *index.Function) types.Type {
	if f == nil || f.Return != "" || f.DocReturn != "" || f.File != e.File.Path {
		return types.Unknown
	}
	return e.bodyReturn(f.Span)
}

// methodBodyReturn infers the return type of method m of class cls from
// its body when m is declared in this file without declared or documented
// return type and the call cannot dispatch to an override: m is private or
// final, its class final or an enum, or the call is non-virtual (`self::`,
// `parent::`). Trait methods are skipped ($this/self mean the user class).
func (e *Env) methodBodyReturn(m *index.Method, virtual bool) types.Type {
	if m == nil || m.Return != "" || m.DocReturn != "" || m.Abstract {
		return types.Unknown
	}
	c := e.Index.Class(m.Class, e.PHP)
	if c == nil || c.File != e.File.Path || c.Kind == syntax.KindTrait || c.Kind == syntax.KindInterface {
		return types.Unknown
	}
	if virtual && !m.Final && m.Visibility != index.Private && !c.Final && c.Kind != syntax.KindEnum {
		return types.Unknown
	}
	return e.bodyReturn(m.Span)
}

// bodyReturn infers the return type of the function or method declared at
// span in this file from its `return` statements (nested closures/classes
// excluded): the union of the returned expressions' types, plus null for a
// bare `return;` or when the end of the body is reachable. Unknown for
// generators, bodies that never return normally, when any returned
// expression's type is unknown, on recursion (the body is being inferred)
// and beyond maxBodyDepth nested inferences. Results are cached.
func (e *Env) bodyReturn(span syntax.Span) types.Type {
	if t, ok := e.bodies[span]; ok {
		return t
	}
	if e.bodyBusy[span] || e.bodyDepth >= maxBodyDepth {
		return types.Unknown
	}
	var body *syntax.Block
	switch d := e.declAt(span).(type) {
	case *syntax.Function:
		body = d.Body
	case *syntax.Method:
		body = d.Body
	}
	if body == nil {
		return types.Unknown
	}
	e.bodyBusy[span] = true
	e.bodyDepth++
	t := e.inferBody(body)
	e.bodyDepth--
	delete(e.bodyBusy, span)
	e.bodies[span] = t
	return t
}

func (e *Env) inferBody(body *syntax.Block) types.Type {
	var ts []types.Type
	ok := true
	syntax.Inspect(body, func(n syntax.Node) bool {
		if !ok {
			return false
		}
		switch n := n.(type) {
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
		case *syntax.Yield, *syntax.YieldFrom:
			ok = false
		case *syntax.Return:
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
	if !syntax.Terminates(body) {
		ts = append(ts, types.Null) // falls off the end
	}
	if len(ts) == 0 {
		return types.Unknown // always throws/exits
	}
	return types.Union(ts...)
}

// declAt returns the function or method declared at span in this file.
func (e *Env) declAt(span syntax.Span) syntax.Node {
	if e.decls == nil {
		e.decls = map[syntax.Span]syntax.Node{}
		syntax.InspectFile(e.File, func(n syntax.Node) bool {
			switch n.(type) {
			case *syntax.Function, *syntax.Method:
				e.decls[n.Span()] = n
			}
			return true
		})
	}
	return e.decls[span]
}

// isVirtualClassRef reports whether a static call through class reference x
// may dispatch to an override: `static::` and expressions (`$obj::`), not
// `self::`, `parent::` or a class name.
func isVirtualClassRef(x syntax.Expr) bool {
	n, ok := x.(*syntax.Name)
	return !ok || strings.EqualFold(n.Value, "static")
}

// LiteralTextType infers the type of a constant/default initialiser from its
// source text (literals, arrays, true/false/null); unknown otherwise.
func LiteralTextType(v string) types.Type { return literalTextType(v) }
