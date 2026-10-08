package infer

import (
	"strings"

	"custos/internal/syntax"
	"custos/internal/types"
)

// Closures and callables: a closure or arrow function is `\Closure`
// carrying its return type (declared, documented, else inferred from its
// body like a function's), as are `callable(…): R` / `Closure(…): R` doc
// types; calling such a value (`$f()`, `call_user_func($f)`, an invokable
// object) gives that type, and array_map() maps the callback's return type
// over the elements.

// closureType is the type of closure or arrow function n.
func (e *Env) closureType(n syntax.Expr) types.Type {
	return types.WithCallableReturn(types.Of(`\Closure`), `\Closure`, e.closureReturn(n))
}

// closureReturn is the type a call to closure or arrow function n returns:
// its declared return type (refined by an `@return` doc), else the type
// inferred from its body (unknown for generators, on recursion and beyond
// maxBodyDepth nested inferences). `void` reads as null (the call's value).
func (e *Env) closureReturn(n syntax.Expr) types.Type {
	var retNode syntax.Expr
	var body func() types.Type
	switch c := n.(type) {
	case *syntax.Closure:
		retNode, body = c.ReturnType, func() types.Type { return e.inferBody(c.Body) }
	case *syntax.ArrowFunction:
		retNode, body = c.ReturnType, func() types.Type { return e.TypeOf(c.Expr) }
	}
	at := n.Span().Start
	t := types.FromNode(retNode, e.resolver(at))
	if d := e.DocOf(n); d != nil && !e.native {
		if doc := d.ReturnType(); doc != "" {
			t = pickMemberType(t, types.FromDoc(doc, e.resolverFor(n, at)))
		}
	}
	if t.IsUnknown() {
		span := n.Span()
		if e.bodyBusy[span] || e.bodyDepth >= maxBodyDepth {
			return types.Unknown
		}
		e.bodyBusy[span] = true
		e.bodyDepth++
		t = body()
		e.bodyDepth--
		delete(e.bodyBusy, span)
	}
	t = bindStatic(t, e.selfClass(syntax.EnclosingClass(n)))
	return voidAsNull(t)
}

// voidAsNull replaces `void` by `null`: the value a call evaluates to.
func voidAsNull(t types.Type) types.Type {
	switch {
	case !t.Has("void"):
		return t
	case t.OnlyOf("void"):
		return types.Null
	}
	return types.Union(t.Without("void"), types.Null)
}

// invokeType is the type a call to a value of type t returns: the return
// type a callable signature or closure gives its `callable`/`\Closure`
// members, `__invoke()` for objects. Unknown as soon as one member is not
// understood (strings and arrays naming callables, a closure of unknown
// return type); null and false (calling them fails) are ignored.
func (e *Env) invokeType(t types.Type) types.Type {
	callee := t.Without("null", "false")
	if callee.IsUnknown() {
		return types.Unknown
	}
	var ts []types.Type
	var objects []string
	sig := false
	for _, a := range callee.Atoms() {
		switch {
		case a == "callable" || strings.EqualFold(a, `\Closure`):
			sig = true
		case strings.HasPrefix(a, `\`) && !strings.HasSuffix(a, "[]"):
			objects = append(objects, a)
		default:
			return types.Unknown
		}
	}
	if sig {
		r := types.CallableReturn(callee)
		if r.IsUnknown() {
			return types.Unknown
		}
		ts = append(ts, r)
	}
	for _, a := range objects {
		cls := strings.TrimPrefix(a, `\`)
		r := e.methodReturn(cls, "__invoke", true, cls, callee.TypeArgs(a), nil)
		if r.IsUnknown() {
			return types.Unknown
		}
		ts = append(ts, voidAsNull(r))
	}
	return types.Union(ts...)
}

// callbackReturn is the type a callback argument x returns when called:
// see invokeType, plus a string literal naming a global function
// (`array_map('trim', $xs)`).
func (e *Env) callbackReturn(x syntax.Expr) types.Type {
	if lit, ok := syntax.UnwrapParens(x).(*syntax.Literal); ok && lit.LitKind == syntax.LitString {
		name, ok := plainString(lit.Raw)
		if !ok || name == "" || strings.Contains(name, ":") {
			return types.Unknown
		}
		f := e.Index.Function(strings.TrimPrefix(name, `\`), e.PHP)
		if f == nil || f.Tpl != nil {
			return types.Unknown
		}
		if e.userDoc(f.Builtin) {
			return voidAsNull(types.FromDoc(f.Return, nil))
		}
		if f.Return == "" && f.DocReturn == "" {
			return voidAsNull(e.BodyReturnType(f))
		}
		if f.Builtin {
			return voidAsNull(builtinMemberType(f.Return, f.DocReturn))
		}
		return voidAsNull(memberType(f.Return, f.DocReturn))
	}
	return e.invokeType(e.TypeOf(x))
}

// arrayOf is the array type with elements of type el (`T[]` for each
// member of el), keeping the element's array facts.
func arrayOf(el types.Type) types.Type {
	atoms := make([]string, 0, len(el.Atoms()))
	for _, a := range el.Atoms() {
		atoms = append(atoms, a+"[]")
	}
	return types.Of(atoms...).WithElem(el)
}

// arrayMapType types `array_map($cb, $xs, …)`: an array of the callback's
// return type (keys are kept or renumbered, the count is not changed, but
// neither fact is recorded); ok is false when that type is unknown or
// mixed, or the callback is null (array_map then zips the arrays).
func (e *Env) arrayMapType(cb syntax.Expr) (types.Type, bool) {
	r := e.callbackReturn(cb)
	if r.IsUnknown() || r.Has("mixed") || r.Has("void") {
		return types.Unknown, false
	}
	return arrayOf(r), true
}

// arrayFilterType types `array_filter($xs[, $cb])`: the elements of $xs
// (a sealed shape's values) and, without a callback, minus null and false
// (array_filter() then drops the falsy values). Keys are kept but gaps may
// appear and the result may be empty: no shape and no non-empty fact.
func (e *Env) arrayFilterType(arr syntax.Expr, at types.Type, withCallback bool) types.Type {
	el := e.elemOf(at, arr)
	if el.IsUnknown() || el.Has("mixed") {
		return at.WithoutShape().WithNonEmpty(false)
	}
	if !withCallback {
		if kept := el.Without("null", "false"); len(kept.Atoms()) > 0 {
			el = kept
		}
	}
	return arrayOf(el)
}
