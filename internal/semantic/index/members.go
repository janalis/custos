package index

import (
	"strings"

	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
)

func (x *extractor) classBody(n *syntax.ClassLike) {
	at := n.Span().Start
	fqn := x.r.SymbolFQN(n)
	c := &Class{
		Anonymous: n.Name == nil,
		FQN:       fqn, Kind: n.ClassKind, Abstract: n.Modifiers.Has(syntax.TAbstract), Final: n.Modifiers.Has(syntax.TFinal),
		Readonly: n.Modifiers.Has(syntax.TReadonly), Methods: map[string]*Method{}, Props: map[string]*Property{},
		Consts: map[string]*ClassConst{}, File: x.f.Path, Span: n.Span(), Avail: x.avail(n.Attrs, nil),
	}
	switch n.ClassKind {
	case syntax.KindInterface:
		for _, e := range n.Extends {
			c.Interfaces = append(c.Interfaces, x.r.Class(e.Value, at))
		}
	default:
		if len(n.Extends) > 0 {
			c.Parent = x.r.Class(n.Extends[0].Value, at)
		}
	}
	for _, i := range n.Implements {
		c.Interfaces = append(c.Interfaces, x.r.Class(i.Value, at))
	}
	for _, g := range n.Attrs {
		for _, a := range g.Attrs {
			c.Attrs = append(c.Attrs, x.r.Class(a.Name.Value, a.Span().Start))
		}
	}
	if d := x.doc(n); d != nil {
		c.Deprecated = d.Has("deprecated")
		c.Avail = x.avail(n.Attrs, d)
		x.magicMembers(c, d, at)
		x.generics(c, d, at)
	}
	for _, m := range n.Members {
		switch m := m.(type) {
		case *syntax.Method:
			x.method(c, m)
		case *syntax.Property:
			d := x.doc(m)
			for _, p := range m.Props {
				prop := &Property{
					Name: p.Var.Name, Class: fqn, Visibility: visibility(m.Modifiers), Static: m.Modifiers.Has(syntax.TStatic),
					SetVisibility: setVisibility(m.Modifiers), Final: m.Modifiers.Has(syntax.TFinal) || m.Modifiers.Has(syntax.TPrivateSet),
					Readonly: m.Modifiers.Has(syntax.TReadonly) || c.Readonly, Type: x.typeStr(m.Type, at), HasDefault: p.Default != nil,
					Default: x.text(p.Default), Span: p.Span(), ReadsRunCode: readsRunCode(p.Var.Name, m.Modifiers, m.Hooks),
					WritesRunCode: writesRunCode(p.Var.Name, m.Modifiers, m.Hooks),
					Hooked:        len(m.Hooks) > 0, Attributed: len(m.Attrs) > 0,
				}
				if d != nil {
					prop.DocType = x.docTypeStr(d.EffectiveVarType(p.Var.Name), at)
				}
				c.Props[prop.Name] = prop
			}
		case *syntax.ClassConst:
			for _, k := range m.Consts {
				c.Consts[k.Name.Value] = &ClassConst{
					Name: k.Name.Value, Class: fqn, Visibility: visibility(m.Modifiers),
					Type:  x.typeStr(m.Type, at),
					Final: m.Modifiers.Has(syntax.TFinal), Value: x.text(k.Value), Span: k.Span(),
				}
			}
		case *syntax.EnumCase:
			c.Consts[m.Name.Value] = &ClassConst{Name: m.Name.Value, Class: fqn, Case: true, Value: x.text(m.Value), Span: m.Span()}
		case *syntax.TraitUse:
			for _, t := range m.Traits {
				c.Traits = append(c.Traits, x.r.Class(t.Value, m.Span().Start))
			}
			for _, a := range m.Adaptations {
				adapt := TraitAdaptation{Method: a.Method.Value}
				if a.Trait != nil {
					adapt.Trait = x.r.Class(a.Trait.Value, a.Span().Start)
				}
				for _, excluded := range a.Insteadof {
					adapt.Insteadof = append(adapt.Insteadof, x.r.Class(excluded.Value, a.Span().Start))
				}
				if a.Alias != nil {
					adapt.Alias = a.Alias.Value
				}
				if a.Modifier != nil {
					switch a.Modifier.Kind {
					case syntax.TPublic, syntax.TProtected, syntax.TPrivate:
						v := visibility(syntax.Modifiers{*a.Modifier})
						adapt.Visibility = &v
					case syntax.TFinal:
						adapt.Final = true
					}
				}
				c.TraitAdaptations = append(c.TraitAdaptations, adapt)
			}
		}
	}
	if n.ClassKind == syntax.KindEnum {
		addEnumMembers(c, x.typeStr(n.EnumType, at))
	}
	x.out.Classes = append(x.out.Classes, c)
}

// readsRunCode reports whether reading the property `name` declared with
// these modifiers and hooks may run code: it is abstract, has a `get` hook
// or an abstract hook, or is virtual (no hook touches its backing store).
func readsRunCode(name string, mods syntax.Modifiers, hooks []*syntax.PropertyHook) bool {
	if mods.Has(syntax.TAbstract) {
		return true
	}
	if len(hooks) == 0 {
		return false
	}
	backed := false
	for _, h := range hooks {
		if h.Body == nil || strings.EqualFold(h.Name.Value, "get") {
			return true
		}
		if _, short := h.Body.(syntax.Expr); short || refsBackingStore(h.Body, name) {
			backed = true // `set => expr` assigns the backing store
		}
	}
	return !backed
}

// writesRunCode is conservative for abstract and virtual properties, and
// records any set hook even when its body only assigns the backing store.
func writesRunCode(name string, mods syntax.Modifiers, hooks []*syntax.PropertyHook) bool {
	if mods.Has(syntax.TAbstract) {
		return true
	}
	if len(hooks) == 0 {
		return false
	}
	backed := false
	for _, h := range hooks {
		if h.Body == nil || strings.EqualFold(h.Name.Value, "set") {
			return true
		}
		if refsBackingStore(h.Body, name) {
			backed = true
		}
	}
	return !backed
}

func setVisibility(mods syntax.Modifiers) *Visibility {
	var vis Visibility
	switch {
	case mods.Has(syntax.TPrivateSet):
		vis = Private
	case mods.Has(syntax.TProtectedSet):
		vis = Protected
	case mods.Has(syntax.TPublicSet):
		vis = Public
	default:
		return nil
	}
	return &vis
}

// refsBackingStore reports whether body reads or writes `$this->name`.
func refsBackingStore(body syntax.Node, name string) bool {
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if f, ok := n.(*syntax.PropertyFetch); ok {
			v, isVar := f.Var.(*syntax.Variable)
			id, isID := f.Name.(*syntax.Identifier)
			if isVar && isID && v.Name == "this" && id.Value == name {
				found = true
			}
		}
		return !found
	})
	return found
}

func visibility(m syntax.Modifiers) Visibility {
	switch {
	case m.Has(syntax.TPrivate):
		return Private
	case m.Has(syntax.TProtected):
		return Protected
	}
	return Public
}

func (x *extractor) params(ps []*syntax.Param, d *phpdoc.Doc, at uint32) []Param {
	docTypes := map[string]string{}
	outTypes := map[string]string{}
	if d != nil {
		for _, p := range d.EffectiveParams() {
			docTypes[p.Name] = p.Type
		}
		for _, tag := range []string{"param-out", "psalm-param-out", "phpstan-param-out"} { // later wins
			for _, p := range d.ParamsOf(tag) {
				if p.Name != "" && p.Type != "" {
					outTypes[p.Name] = p.Type
				}
			}
		}
	}
	out := make([]Param, 0, len(ps))
	for _, p := range ps {
		prm := Param{
			Name: p.Var.Name, Type: x.typeStr(p.Type, at), DocType: x.docTypeStr(docTypes[p.Var.Name], at),
			Optional: p.Default != nil || p.Variadic, ByRef: p.ByRef, Variadic: p.Variadic,
			Promoted: len(p.Modifiers) > 0, Default: x.text(p.Default), Avail: x.avail(p.Attrs, nil),
		}
		if p.ByRef {
			prm.Out = x.docTypeStr(outTypes[p.Var.Name], at)
		}
		out = append(out, prm)
	}
	return out
}

func (x *extractor) method(c *Class, m *syntax.Method) {
	d := x.doc(m)
	saved := x.methodTpl
	x.methodTpl = nil
	if d != nil {
		for _, t := range d.Templates() {
			if x.methodTpl == nil {
				x.methodTpl = map[string]bool{}
			}
			x.methodTpl[t] = true
		}
	}
	x.withTemplates(d, func() { x.methodBody(c, m, d) })
	x.methodTpl = saved
}

// voidDoc drops a documented `void` return contradicted by the body: a
// `return value;` of the function itself (`@return void` left over after
// a refactoring); the body's own type then applies.
func voidDoc(doc string, body *syntax.Block) string {
	if !strings.EqualFold(doc, "void") || body == nil {
		return doc
	}
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
			return false
		case *syntax.Return:
			found = found || n.Expr != nil
		}
		return !found
	})
	if found {
		return ""
	}
	return doc
}

// callsCSPRNG reports whether body calls a secure random generator by name
// (a function or a method of that name, as the IV rule's same-file check).
func callsCSPRNG(body syntax.Node) bool {
	found := false
	syntax.Inspect(body, func(n syntax.Node) bool {
		if found {
			return false
		}
		var name string
		switch c := n.(type) {
		case *syntax.FuncCall:
			if nm, ok := c.Name.(*syntax.Name); ok {
				name = nm.Value[strings.LastIndexByte(nm.Value, '\\')+1:]
			}
		case *syntax.MethodCall:
			if id, ok := c.Name.(*syntax.Identifier); ok {
				name = id.Value
			}
		case *syntax.StaticCall:
			if id, ok := c.Name.(*syntax.Identifier); ok {
				name = id.Value
			}
		}
		switch strings.ToLower(name) {
		case "random_bytes", "openssl_random_pseudo_bytes", "mcrypt_create_iv":
			found = true
		}
		return !found
	})
	return found
}

// storedParams reports whether constructor m only stores its parameters
// (each statement `$this->prop = $param;`, promoted parameters aside) and
// the properties it sets that way.
func storedParams(m *syntax.Method) (bool, []string) {
	if m.Body == nil {
		return false, nil
	}
	params := map[string]bool{}
	var props []string
	for _, p := range m.Params {
		params[p.Var.Name] = true
		if len(p.Modifiers) > 0 {
			props = append(props, p.Var.Name)
		}
	}
	for _, st := range m.Body.Stmts {
		es, ok := st.(*syntax.ExprStmt)
		if !ok {
			return false, nil
		}
		a, ok := es.Expr.(*syntax.Assign)
		if !ok || a.Op.Kind != syntax.TEqual || a.ByRef {
			return false, nil
		}
		pf, ok := a.Var.(*syntax.PropertyFetch)
		if !ok || pf.NullSafe {
			return false, nil
		}
		this, ok := pf.Var.(*syntax.Variable)
		id, ok2 := pf.Name.(*syntax.Identifier)
		v, ok3 := syntax.UnwrapParens(a.Value).(*syntax.Variable)
		if !ok || !ok2 || !ok3 || this.Name != "this" || v.NameExpr != nil || !params[v.Name] {
			return false, nil
		}
		props = append(props, id.Value)
	}
	return true, props
}

// promotes reports whether a constructor parameter list promotes a
// property (a modifier on a parameter).
func promotes(ps []*syntax.Param) bool {
	for _, p := range ps {
		if len(p.Modifiers) > 0 {
			return true
		}
	}
	return false
}

func (x *extractor) methodBody(c *Class, m *syntax.Method, d *phpdoc.Doc) {
	at := m.Span().Start
	meth := &Method{
		Name: m.Name.Value, Class: c.FQN, Visibility: visibility(m.Modifiers), Static: m.Modifiers.Has(syntax.TStatic),
		Abstract: m.Modifiers.Has(syntax.TAbstract) || c.Kind == syntax.KindInterface, Final: m.Modifiers.Has(syntax.TFinal),
		ByRef: m.ByRef, Params: x.params(m.Params, d, at), Span: m.Span(),
		Avail: x.avail(m.Attrs, d), EmptyBody: m.Body != nil && len(m.Body.Stmts) == 0 && !promotes(m.Params),
	}
	if strings.EqualFold(m.Name.Value, "__construct") {
		meth.StoresParams, meth.Stores = storedParams(m)
	}
	if m.Body != nil {
		meth.CSPRNG = callsCSPRNG(m.Body)
	}
	meth.Return, meth.RetVer = x.returnType(m.ReturnType, m.Attrs, at)
	if d != nil {
		meth.DocReturn = voidDoc(x.docTypeStr(d.EffectiveReturnType(), at), m.Body)
		meth.GenReturn = x.genReturn(d, at)
		meth.CondReturn = x.condReturn(d, at)
		meth.Asserts = x.assertions(d, m.Params, !meth.Static, at)
		meth.Tpl = x.funcTemplates(d, m.Params, meth.Asserts, at)
		meth.Deprecated = d.Has("deprecated")
	}
	c.Methods[strings.ToLower(meth.Name)] = meth
	if strings.EqualFold(meth.Name, "__construct") {
		for i, p := range m.Params {
			if len(p.Modifiers) == 0 {
				continue
			}
			// The constructor's `@param Foo[] $items` documents the
			// promoted property too (as PhpStorm and PHPStan read it).
			c.Props[p.Var.Name] = &Property{
				Name: p.Var.Name, Class: c.FQN, Visibility: visibility(p.Modifiers),
				SetVisibility: setVisibility(p.Modifiers), Final: p.Modifiers.Has(syntax.TFinal) || p.Modifiers.Has(syntax.TPrivateSet),
				Readonly: p.Modifiers.Has(syntax.TReadonly) || c.Readonly, Type: x.typeStr(p.Type, at), Promoted: true,
				DocType: meth.Params[i].DocType, HasDefault: p.Default != nil, Default: x.text(p.Default), Span: p.Span(),
				ReadsRunCode:  readsRunCode(p.Var.Name, p.Modifiers, p.Hooks),
				WritesRunCode: writesRunCode(p.Var.Name, p.Modifiers, p.Hooks),
				Hooked:        len(p.Hooks) > 0, Attributed: len(p.Attrs) > 0,
			}
		}
	}
}
