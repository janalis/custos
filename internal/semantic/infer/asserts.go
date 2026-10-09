package infer

import (
	"strings"

	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/semantic/types"
)

// Assertion annotations: `@phpstan-assert T $x` (and `@psalm-assert`) on a
// function or method narrows the argument passed for $x after a call
// statement; `-assert-if-true` / `-assert-if-false` narrow it where the
// call, used as a condition, is true / false (if/elseif/while bodies,
// ternaries, && / || operands, early-exit guards). Targets may also be the
// receiver (`$this`) or one of its properties (`$this->prop`, narrowed when
// the call is made on `$this`). `!T` removes T; `=T` reads as T. Template
// types bind from the call's arguments (`assertInstanceOf(Foo::class, $x)`).
// As other guards, an assertion never makes an unknown type known.

// callAsserts is the callee information needed to apply the assertions of
// one call expression.
type callAsserts struct {
	asserts []index.Assertion
	params  []index.Param
	ft      *index.FuncTemplates
	cls     *index.Class // declaring class (methods)
	args    *syntax.ArgList
	recv    syntax.Expr // receiver of a method call; nil otherwise
	// origin and recvArgs bind the declaring class's templates.
	origin   string
	recvArgs []types.Type

	bound  bool // mb and classB computed
	mb     tplBindings
	classB tplBindings
	typs   []types.Type // asserted types by index in asserts (lazy; unknown: unusable)
	done   []bool
}

// assertsOf returns the assertions of the function or method called by x
// (nil when x is no call or its callee declares none). Cached per node.
func (e *Env) assertsOf(x syntax.Expr) *callAsserts {
	switch x.(type) {
	case *syntax.FuncCall, *syntax.MethodCall, *syntax.StaticCall:
	default:
		return nil
	}
	if ca, ok := e.asserts[x]; ok {
		return ca
	}
	ca := e.resolveAsserts(x)
	if e.asserts == nil {
		e.asserts = map[syntax.Expr]*callAsserts{}
	}
	e.asserts[x] = ca
	return ca
}

func (e *Env) resolveAsserts(x syntax.Expr) *callAsserts {
	switch n := x.(type) {
	case *syntax.FuncCall:
		if isFirstClassCallable(n.Args) {
			return nil
		}
		f := e.ResolveFunction(n)
		if f == nil || len(f.Asserts) == 0 || e.userDoc(f.Builtin) {
			return nil
		}
		return &callAsserts{asserts: f.Asserts, params: f.Params, ft: f.Tpl, args: n.Args}
	case *syntax.MethodCall:
		id, ok := n.Name.(*syntax.Identifier)
		if !ok || isFirstClassCallable(n.Args) {
			return nil
		}
		recv := e.TypeOf(n.Var)
		var m *index.Method
		var origin string
		for _, cls := range e.memberClasses(recv, func(c string) bool { return e.Index.FindMethod(c, id.Value, e.PHP) != nil }) {
			c := strings.TrimPrefix(cls, `\`)
			cm := e.Index.FindMethod(c, id.Value, e.PHP)
			if cm == nil || (m != nil && cm != m) {
				return nil // unknown, or different methods: no assertion
			}
			m, origin = cm, c
		}
		if m == nil || len(m.Asserts) == 0 || e.userDoc(m.Builtin) {
			return nil
		}
		return &callAsserts{
			asserts: m.Asserts, params: m.Params, ft: m.Tpl, cls: e.Index.Class(m.Class, e.PHP),
			args: n.Args, recv: n.Var, origin: origin, recvArgs: recv.TypeArgs(`\` + origin),
		}
	}
	n := x.(*syntax.StaticCall) // assertsOf admits only the three call kinds
	id, ok := n.Name.(*syntax.Identifier)
	if !ok || isFirstClassCallable(n.Args) {
		return nil
	}
	cls := e.classRef(n.Class)
	if cls == "" {
		return nil
	}
	m := e.Index.FindMethod(cls, id.Value, e.PHP)
	if m == nil || len(m.Asserts) == 0 || e.userDoc(m.Builtin) {
		return nil
	}
	ca := &callAsserts{asserts: m.Asserts, params: m.Params, ft: m.Tpl, cls: e.Index.Class(m.Class, e.PHP), args: n.Args, origin: cls}
	if !m.Static {
		// `self::isFoo()` / `parent::isFoo()` from an instance method
		// calls it on $this.
		if nm, ok := n.Class.(*syntax.Name); ok {
			switch strings.ToLower(nm.Value) {
			case "self", "static", "parent":
				ca.recv = &syntax.Variable{Name: "this"}
			}
		}
	}
	return ca
}

// applyAsserts applies the assertions of kind in ca whose target is the
// expression identified by key name (see narrowKey) to its type t, negated
// when negate is set (an -if-true assertion in the false branch). ok
// reports whether one applied.
func (e *Env) applyAsserts(ca *callAsserts, kind index.AssertKind, name string, t types.Type, negate bool) (types.Type, bool) {
	if ca == nil || t.IsUnknown() {
		return t, false
	}
	applied := false
	for i, a := range ca.asserts {
		if a.Kind != kind || !e.assertTargets(ca, a, name) {
			continue
		}
		if ca.typs == nil {
			ca.typs, ca.done = make([]types.Type, len(ca.asserts)), make([]bool, len(ca.asserts))
		}
		if !ca.done[i] {
			ca.typs[i], _ = e.assertedType(ca, a)
			ca.done[i] = true
		}
		at := ca.typs[i]
		if at.IsUnknown() {
			continue
		}
		if a.Negated != negate {
			t = assertNot(t, at)
		} else {
			t = e.assertIs(t, at)
		}
		applied = true
	}
	return t, applied
}

// assertTargets reports whether assertion a of the call designates name.
func (e *Env) assertTargets(ca *callAsserts, a index.Assertion, name string) bool {
	switch a.Param {
	case index.AssertThis:
		return ca.recv != nil && narrowKey(syntax.UnwrapParens(ca.recv)) == name
	case index.AssertThisProp:
		return ca.recv != nil && narrowKey(syntax.UnwrapParens(ca.recv)) == "this" && name == "this->"+a.Prop
	}
	arg := tplArg(ca.args, ca.params, a.Param)
	return arg != nil && isVar(arg, name)
}

// assertedType is the type of assertion a, with templates bound.
func (e *Env) assertedType(ca *callAsserts, a index.Assertion) (types.Type, bool) {
	if !strings.Contains(a.Type, `\~`) {
		t := types.FromDoc(a.Type, nil)
		return t, !t.IsUnknown()
	}
	if !ca.bound {
		ca.bound = true
		ca.mb = e.bindCall(ca.ft, ca.params, ca.args)
		if ca.cls != nil && len(ca.cls.Templates) > 0 && ca.origin != "" {
			ca.classB = e.genBindings(ca.origin, ca.recvArgs)[strings.ToLower(ca.cls.FQN)]
		}
	}
	t, ok := substFuncTemplates(a.Type, ca.mb, ca.classB, ca.cls)
	if !ok || t.Has("mixed") {
		return types.Unknown, false
	}
	return t, true
}

// assertIs narrows t to the asserted type a: the members of t that belong
// to a (a subclass of an asserted class, `true` for `bool`, arrays for
// `iterable`…), or a itself when none does.
// (The index drops `mixed` assertions and assertedType rejects bindings
// to mixed, so a never contains mixed.)
func (e *Env) assertIs(t, a types.Type) types.Type {
	var drop []string
	for _, x := range t.Atoms() {
		if !e.atomIn(x, a) {
			drop = append(drop, x)
		}
	}
	if len(drop) == len(t.Atoms()) {
		return a
	}
	return t.Without(drop...)
}

// atomIn reports whether atom x of a value's type is part of type a.
func (e *Env) atomIn(x string, a types.Type) bool {
	switch {
	case a.Has(x):
		return true
	case x == "true" || x == "false":
		return a.Has("bool")
	case strings.HasSuffix(x, "[]") || x == "array":
		return a.Has("iterable") || (x != "array" && a.Has("array"))
	case strings.HasPrefix(x, `\`):
		return a.Has("object") || e.subtypeOfAny(strings.TrimPrefix(x, `\`), a.Classes())
	}
	return false
}

// assertNot removes the asserted type a from t (`!null`, `!Foo`); t is
// kept when nothing would remain.
func assertNot(t, a types.Type) types.Type {
	out := t.Without(a.Atoms()...)
	switch {
	case a.Has("bool"):
		out = out.Without("true", "false")
	case out.Has("bool") && a.Has("true") != a.Has("false"):
		// bool without true is false (and conversely).
		other := "true"
		if a.Has("true") {
			other = "false"
		}
		out = replaceAtom(out, "bool", other)
	}
	if len(out.Atoms()) == 0 {
		return t
	}
	return out
}
