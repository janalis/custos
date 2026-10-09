package infer

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// ResolveFunction finds the index entry for a named call (nil for dynamic calls).
func (e *Env) ResolveFunction(call *syntax.FuncCall) *index.Function {
	name, ok := call.Name.(*syntax.Name)
	if !ok {
		return nil
	}
	fqn, fb := e.Names.Function(name.Value, name.Span().Start)
	f := e.Index.ResolveFunction(fqn, fb, e.PHP)
	if e.annotating && fb != "" && f != nil && !strings.EqualFold(f.FQN, fqn) {
		e.deps = append(e.deps, fqn)
	}
	return f
}

func (e *Env) funcCallType(n *syntax.FuncCall) types.Type {
	name, named := n.Name.(*syntax.Name)
	if !named {
		return e.invokeType(e.TypeOf(n.Name)) // `$f()`, `(fn() => 1)()`
	}
	if t, ok := e.overrideType(n, name); ok {
		return t
	}
	f := e.ResolveFunction(n)
	if f == nil {
		return types.Unknown
	}
	t := e.declCallType(f, n)
	if st, ok := e.safeCallType(f, n, t); ok {
		return st
	}
	if decls := e.Index.FunctionDecls(f.FQN, e.PHP); len(decls) > maxFuncDecls {
		return types.Unknown // hostile: thousands of declarations of one name
	} else if len(decls) > 1 {
		// Declared more than once (a no-op variant loaded instead of the
		// real one): any declaration may be the one that runs.
		ts := []types.Type{t}
		for _, g := range decls {
			if g != f {
				ts = append(ts, e.declCallType(g, n))
			}
		}
		t = types.Union(ts...)
	}
	return t
}

// safeCallType types a call to thecodingmachine/safe's `Safe\X`, which
// wraps the builtin X and throws where X returns false (and, for the preg_
// functions, null): the builtin's type for these arguments without false
// (and null), unless the Safe declaration's own type t is narrower.
func (e *Env) safeCallType(f *index.Function, n *syntax.FuncCall, t types.Type) (types.Type, bool) {
	ns, base, ok := strings.Cut(f.FQN, `\`)
	if !ok || !strings.EqualFold(ns, "safe") || strings.Contains(base, `\`) {
		return types.Unknown, false
	}
	b := e.Index.Function(base, e.PHP)
	if b == nil || !b.Builtin {
		return types.Unknown, false
	}
	bt, ok := e.overrideType(n, &syntax.Name{Value: `\` + base})
	if !ok {
		bt = e.declCallType(b, n)
	}
	drop := []string{"false"}
	if strings.HasPrefix(strings.ToLower(base), "preg_") || (!t.IsUnknown() && !t.Has("null")) {
		drop = append(drop, "null")
	}
	if nt := bt.Without(drop...); len(nt.Atoms()) > 0 {
		bt = nt
	}
	if bt.IsUnknown() {
		return types.Unknown, false
	}
	if f.Return == "" && f.DocReturn == "" {
		return bt, true // the wrapper's body says nothing about its contract
	}
	if !t.IsUnknown() && !t.Has("mixed") {
		for _, a := range bt.Atoms() {
			if !e.atomIn(a, t) {
				return types.Unknown, false // the Safe declaration knows better
			}
		}
	}
	return bt, true
}

// maxFuncDecls caps the declarations of one function a call unions (see
// funcCallType); beyond it the call is unknown.
const maxFuncDecls = 16

// declCallType types call n to function declaration f.
func (e *Env) declCallType(f *index.Function, n *syntax.FuncCall) types.Type {
	if e.userDoc(f.Builtin) {
		return types.FromDoc(f.Return, nil)
	}
	if f.Tpl != nil {
		if t, ok := e.tplReturn(f.Tpl, f.Params, n.Args, f.Return, nil, nil); ok {
			return t
		}
	}
	if f.CondReturn != "" {
		if t, ok := e.condCall(f.CondReturn, f.Params, n.Args, f.Return, f.DocReturn); ok {
			return t
		}
	}
	if f.Return == "" && f.DocReturn == "" {
		return e.BodyReturnType(f)
	}
	if f.Builtin {
		return builtinMemberType(f.Return, f.DocReturn)
	}
	return memberType(f.Return, f.DocReturn)
}
