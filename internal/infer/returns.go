package infer

import (
	"slices"
	"sort"
	"strings"

	"custos/internal/index"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/stubs"
	"custos/internal/syntax"
	"custos/internal/types"
)

// maxBodyDepth caps how many body inferences may be nested (a call in a
// returned expression whose callee's body is inferred, and so on).
const maxBodyDepth = 4

// BodyReturnType infers the return type of a function without declared or
// documented return type from its body: see bodyReturn for a function
// declared in this file; for one declared in another file, the type the
// index computed from its body (AnnotateReturns), if any.
func (e *Env) BodyReturnType(f *index.Function) types.Type {
	if f == nil || f.Return != "" || f.DocReturn != "" {
		return types.Unknown
	}
	if f.File != e.File.Path {
		if f.Inferred == "" || stubs.Index().Function(f.FQN, 0) != nil {
			// A polyfill of a builtin keeps the behaviour it had before
			// (its body is an approximation of the builtin).
			return types.Unknown
		}
		return types.FromDoc(f.Inferred, nil)
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
	if c == nil || c.Kind == syntax.KindTrait || c.Kind == syntax.KindInterface {
		return types.Unknown
	}
	if virtual && !m.Final && m.Visibility != index.Private && !c.Final && c.Kind != syntax.KindEnum {
		return types.Unknown
	}
	if c.File != e.File.Path {
		return types.FromDoc(m.Inferred, nil) // computed by AnnotateReturns
	}
	return e.bodyReturn(m.Span)
}

// AnnotateReturns stores in fs (the symbols extracted from f) the return
// types of its functions and non-private methods that have no declared or
// documented return type, inferred from their bodies as bodyReturn does for
// the current file, but self-contained: the index holds only f's symbols
// over base (the stubs), so anything depending on another file (its
// functions, classes' members, inherited properties) stays unknown. The
// result is used for calls from other files under the same override rules
// as same-file inference (see methodBodyReturn). Namespaced function names
// resolved through the global fallback are recorded in fs.ReturnDeps
// (index.DropStaleInferred).
func AnnotateReturns(f *syntax.File, fs *index.FileSymbols, base *index.Index, php phpver.Version) {
	type target struct {
		span syntax.Span
		set  func(string)
	}
	var todo []target
	for _, fn := range fs.Functions {
		if fn.Return == "" && fn.DocReturn == "" {
			fn := fn
			todo = append(todo, target{fn.Span, func(s string) { fn.Inferred = s }})
		}
	}
	for _, c := range fs.Classes {
		if c.Kind == syntax.KindTrait || c.Kind == syntax.KindInterface {
			continue
		}
		for _, m := range c.Methods {
			// Private methods are only called from their own file; magic
			// @method entries have no body.
			if m.Return != "" || m.DocReturn != "" || m.Abstract || m.Visibility == index.Private || m.Span == (syntax.Span{}) {
				continue
			}
			todo = append(todo, target{m.Span, func(s string) { m.Inferred = s }})
		}
	}
	var props []*index.Property
	for _, c := range fs.Classes {
		for _, p := range c.Props {
			if propEligible(p, c) && p.Span != (syntax.Span{}) {
				props = append(props, p)
			}
		}
	}
	if len(todo) == 0 && len(props) == 0 {
		return
	}
	// Methods come from a map: infer in source order for determinism (the
	// cache and depth limit make results order-dependent at the margins).
	sort.Slice(todo, func(i, j int) bool { return todo[i].span.Start < todo[j].span.Start })
	ix := index.New(base)
	ix.Add(fs)
	e := NewEnv(f, names.New(f), ix, php)
	e.annotating = true
	for _, t := range todo {
		if rt := e.bodyReturn(t.span); !rt.IsUnknown() {
			t.set(rt.DocString())
		}
	}
	// Untyped private properties (see inferredProp), in source order.
	sort.Slice(props, func(i, j int) bool { return props[i].Span.Start < props[j].Span.Start })
	for _, p := range props {
		if c := ix.Class(p.Class, php); c != nil {
			if pt := e.propFromBody(p, c); !pt.IsUnknown() {
				p.Inferred = pt.DocString()
			}
		}
	}
	if len(e.deps) > 0 {
		slices.Sort(e.deps)
		fs.ReturnDeps = slices.Compact(e.deps)
	}
}

// bodyReturn infers the return type of the function or method declared at
// span in this file from its `return` statements (nested closures/classes
// excluded): the union of the returned expressions' types, plus null for a
// bare `return;` or when the end of the body is reachable. Unknown for
// bodies that never return normally, when any returned expression's type
// is unknown, on recursion (the body is being inferred) and beyond
// maxBodyDepth nested inferences. Generators carry their yielded key/value
// types instead of their completion return type. Results are cached.
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
	if t, ok := e.generatorType(body); ok {
		return t
	}
	var ts []types.Type
	ok := true
	syntax.Inspect(body, func(n syntax.Node) bool {
		if !ok {
			return false
		}
		switch n := n.(type) {
		case *syntax.Function, *syntax.Closure, *syntax.ArrowFunction, *syntax.ClassLike:
			return false
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
