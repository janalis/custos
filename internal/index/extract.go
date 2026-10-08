package index

import (
	"sort"
	"strings"

	"custos/internal/names"
	"custos/internal/phpdoc"
	"custos/internal/phpver"
	"custos/internal/syntax"
	"custos/internal/types"
)

type extractor struct {
	f         *syntax.File
	r         *names.Resolver
	out       *FileSymbols
	templates map[string]bool   // @template names in scope (class + member)
	classTpl  map[string]bool   // class-level @template names of the class being extracted
	methodTpl map[string]bool   // @template names of the method being extracted
	aliases   map[string]string // @phpstan-type / import-type names in scope
}

// Extract collects the symbols declared in f.
func Extract(f *syntax.File) *FileSymbols {
	x := &extractor{f: f, r: names.New(f), out: &FileSymbols{Path: f.Path}}
	syntax.InspectFile(f, func(n syntax.Node) bool {
		switch n := n.(type) {
		case *syntax.ClassLike:
			if n.Name != nil {
				x.class(n)
			}
		case *syntax.Function:
			x.function(n)
		case *syntax.ConstStmt:
			for _, c := range n.Consts {
				fqn := c.Name.Value
				if ns := x.r.Namespace(c.Span().Start); ns != "" {
					fqn = ns + `\` + c.Name.Value
				}
				x.out.Constants = append(x.out.Constants, &Constant{FQN: fqn, Value: x.text(c.Value), File: f.Path, Span: c.Span()})
			}
		case *syntax.FuncCall:
			x.define(n)
		}
		return true
	})
	return x.out
}

func (x *extractor) text(n syntax.Node) string {
	if n == nil {
		return ""
	}
	s := n.Span()
	return string(x.f.Src[s.Start:s.End])
}

func (x *extractor) resolver(at uint32) types.Resolver {
	return func(w string) string {
		if x.templates[w] {
			return ""
		}
		if def, ok := x.aliases[w]; ok {
			return "=" + def
		}
		return x.r.Class(w, at)
	}
}

// withTemplates adds the @template names of d for the duration of fn.
func (x *extractor) withTemplates(d *phpdoc.Doc, fn func()) {
	if d == nil {
		fn()
		return
	}
	names := d.Templates()
	if len(names) == 0 {
		fn()
		return
	}
	saved := x.templates
	x.templates = map[string]bool{}
	for k := range saved {
		x.templates[k] = true
	}
	for _, n := range names {
		x.templates[n] = true
	}
	fn()
	x.templates = saved
}

func (x *extractor) typeStr(n syntax.Expr, at uint32) string {
	t := types.FromNode(n, x.resolver(at))
	if t.IsUnknown() {
		return ""
	}
	return t.String()
}

func (x *extractor) docTypeStr(text string, at uint32) string {
	t := types.FromDoc(text, x.resolver(at))
	if t.IsUnknown() {
		return ""
	}
	return t.DocString() // keeps array shapes (round-trips through FromDoc)
}

// DocComment returns the doc comment directly preceding node n ("" if none).
func DocComment(f *syntax.File, n syntax.Node) string {
	start := n.Span().Start
	toks := f.Tokens
	i := sort.Search(len(toks), func(i int) bool { return toks[i].Start >= start })
	for j := i - 1; j >= 0; j-- {
		t := toks[j]
		switch t.Kind {
		case syntax.TWhitespace, syntax.TComment:
			continue
		case syntax.TDocComment:
			return string(f.Src[t.Start:t.End])
		}
		break
	}
	// Doc comment between attributes and the declaration.
	s := n.Span()
	for k := i; k < len(toks) && toks[k].Start < s.End; k++ {
		switch toks[k].Kind {
		case syntax.TDocComment:
			return string(f.Src[toks[k].Start:toks[k].End])
		case syntax.TAttribute, syntax.TWhitespace, syntax.TComment, syntax.TString, syntax.TNameQualified,
			syntax.TNameFullyQualified, syntax.TLParen, syntax.TRParen, syntax.TRBracket, syntax.TComma,
			syntax.TConstantEncapsedString, syntax.TLNumber:
			continue
		}
		break
	}
	return ""
}

func (x *extractor) doc(n syntax.Node) *phpdoc.Doc {
	if c := DocComment(x.f, n); c != "" {
		return phpdoc.Parse(c)
	}
	return nil
}

func (x *extractor) class(n *syntax.ClassLike) {
	d := x.doc(n)
	saved, savedTpl := x.aliases, x.classTpl
	x.classTpl = nil
	if d != nil {
		if a := d.TypeAliases(); len(a) > 0 {
			x.aliases = a
		}
		for _, t := range d.Templates() {
			if x.classTpl == nil {
				x.classTpl = map[string]bool{}
			}
			x.classTpl[t] = true
		}
	}
	x.withTemplates(d, func() { x.classBody(n) })
	x.aliases, x.classTpl = saved, savedTpl
}

// genResolver resolves names like resolver, but keeps the class templates
// of the class being extracted as `~T` (FromDoc atoms `\~T`) so that infer
// can bind them per receiver; method templates read as mixed.
func (x *extractor) genResolver(at uint32) types.Resolver {
	return func(w string) string {
		if x.methodTpl[w] {
			return ""
		}
		if x.classTpl[w] {
			return "~" + w
		}
		// x.templates holds no other names here: genResolver runs while
		// a class or a member is extracted, whose templates are exactly
		// classTpl and methodTpl.
		if def, ok := x.aliases[w]; ok {
			return "=" + def
		}
		return x.r.Class(w, at)
	}
}

// generics records the class templates of d and the generic arguments
// given to the parents, interfaces and traits (@extends Base<Foo> …).
func (x *extractor) generics(c *Class, d *phpdoc.Doc, at uint32) {
	for _, p := range d.TemplateParams() {
		c.Templates = append(c.Templates, Template{Name: p.Name, Bound: x.docTypeStr(p.Bound, at)})
	}
	for _, t := range d.Tags {
		switch t.Name {
		case "extends", "template-extends", "phpstan-extends", "psalm-extends",
			"implements", "template-implements", "phpstan-implements", "psalm-implements",
			"use", "template-use", "phpstan-use", "psalm-use":
		default:
			continue
		}
		typ, _ := phpdoc.SplitType(t.Text)
		if !strings.Contains(typ, "<") {
			continue
		}
		st := types.FromDoc(typ, x.genResolver(at))
		cs := st.Classes()
		if len(cs) != 1 {
			continue
		}
		args := st.TypeArgs(cs[0])
		if len(args) == 0 {
			continue
		}
		sa := SuperArgs{Class: strings.TrimPrefix(cs[0], `\`), Args: make([]string, len(args))}
		for i, a := range args {
			sa.Args[i] = "mixed"
			if !a.IsUnknown() {
				sa.Args[i] = a.DocString()
			}
		}
		c.Supers = append(c.Supers, sa)
	}
}

// genReturn returns the documented return type of d when it mentions a
// class template (as `\~T` atoms), preferring @phpstan-return and
// @psalm-return, which carry the generic form when @return does not.
func (x *extractor) genReturn(d *phpdoc.Doc, at uint32) string {
	if len(x.classTpl) == 0 {
		return ""
	}
	for _, tag := range []string{"phpstan-return", "psalm-return", "return"} {
		t, ok := d.Tag(tag)
		if !ok {
			continue
		}
		typ, _ := phpdoc.SplitType(t.Text)
		if ds := types.FromDoc(typ, x.genResolver(at)).DocString(); strings.Contains(ds, `\~`) {
			return ds
		}
	}
	return ""
}

func (x *extractor) classBody(n *syntax.ClassLike) {
	at := n.Span().Start
	ns := x.r.Namespace(at)
	fqn := n.Name.Value
	if ns != "" {
		fqn = ns + `\` + fqn
	}
	c := &Class{
		FQN: fqn, Kind: n.ClassKind, Abstract: n.Modifiers.Has(syntax.TAbstract), Final: n.Modifiers.Has(syntax.TFinal),
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
				prop := &Property{Name: p.Var.Name, Class: fqn, Visibility: visibility(m.Modifiers), Static: m.Modifiers.Has(syntax.TStatic),
					Readonly: m.Modifiers.Has(syntax.TReadonly) || c.Readonly, Type: x.typeStr(m.Type, at), HasDefault: p.Default != nil,
					Default: x.text(p.Default), Span: p.Span()}
				if d != nil {
					prop.DocType = x.docTypeStr(d.VarType(p.Var.Name), at)
				}
				c.Props[prop.Name] = prop
			}
		case *syntax.ClassConst:
			for _, k := range m.Consts {
				c.Consts[k.Name.Value] = &ClassConst{Name: k.Name.Value, Class: fqn, Visibility: visibility(m.Modifiers),
					Final: m.Modifiers.Has(syntax.TFinal), Value: x.text(k.Value), Span: k.Span()}
			}
		case *syntax.EnumCase:
			c.Consts[m.Name.Value] = &ClassConst{Name: m.Name.Value, Class: fqn, Case: true, Value: x.text(m.Value), Span: m.Span()}
		case *syntax.TraitUse:
			for _, t := range m.Traits {
				c.Traits = append(c.Traits, x.r.Class(t.Value, m.Span().Start))
			}
		}
	}
	x.out.Classes = append(x.out.Classes, c)
}

func (x *extractor) magicMembers(c *Class, d *phpdoc.Doc, at uint32) {
	for _, tagName := range []string{"property", "property-read", "property-write"} {
		for _, t := range d.All(tagName) {
			typ, rest := phpdoc.SplitType(t.Text)
			name := phpdoc.VarName(rest)
			if name == "" {
				continue
			}
			if _, exists := c.Props[name]; !exists {
				c.Props[name] = &Property{Name: name, Class: c.FQN, DocType: x.docTypeStr(typ, at), Magic: true, Readonly: tagName == "property-read"}
			}
		}
	}
	for _, t := range d.All("method") {
		text := strings.TrimSpace(t.Text)
		static := false
		if strings.HasPrefix(text, "static ") {
			static, text = true, strings.TrimSpace(text[7:])
		}
		open := strings.IndexByte(text, '(')
		if open < 0 {
			continue
		}
		head := strings.TrimSpace(text[:open])
		ret, name := "", head
		if i := strings.LastIndexAny(head, " \t"); i >= 0 {
			ret, name = head[:i], head[i+1:]
		}
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, exists := c.Methods[key]; !exists {
			c.Methods[key] = &Method{Name: name, Class: c.FQN, Static: static, DocReturn: x.docTypeStr(ret, at)}
		}
	}
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
		for _, p := range d.Params() {
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

// tplResolver resolves names like genResolver, but keeps the templates of
// the function or method being extracted as `~~T` (FromDoc atoms `\~~T`).
func (x *extractor) tplResolver(at uint32) types.Resolver {
	gen := x.genResolver(at)
	return func(w string) string {
		if x.methodTpl[w] {
			return "~~" + w
		}
		return gen(w)
	}
}

// funcTemplates describes the templates of a function or method (the
// names in x.methodTpl) when its documented return type uses them.
// assertions reads the @phpstan-assert / @psalm-assert tags of d (and
// their -if-true / -if-false variants) whose target is a parameter of ps,
// `$this` or `$this->prop` (method only). Types use tplResolver. d is non-nil.
func (x *extractor) assertions(d *phpdoc.Doc, ps []*syntax.Param, method bool, at uint32) []Assertion {
	var out []Assertion
	for _, t := range d.Tags {
		var kind AssertKind
		switch t.Name {
		case "phpstan-assert", "psalm-assert":
			kind = AssertAlways
		case "phpstan-assert-if-true", "psalm-assert-if-true":
			kind = AssertIfTrue
		case "phpstan-assert-if-false", "psalm-assert-if-false":
			kind = AssertIfFalse
		default:
			continue
		}
		typ, rest := phpdoc.SplitType(t.Text)
		a := Assertion{Kind: kind, Param: -3}
		if strings.HasPrefix(typ, "!") {
			a.Negated, typ = true, typ[1:]
		} else {
			typ = strings.TrimPrefix(typ, "=") // `=T`: the same, without negative inference
		}
		name := phpdoc.VarName(rest)
		if name == "" {
			continue
		}
		after := strings.TrimSpace(rest)[1+len(name):]
		switch {
		case name == "this" && method && strings.HasPrefix(after, "->"):
			prop := phpdoc.VarName("$" + after[2:])
			tail := after[2+len(prop):]
			if prop == "" || strings.HasPrefix(strings.TrimSpace(tail), "(") || strings.HasPrefix(tail, "->") || strings.HasPrefix(tail, "[") {
				continue // a method call or a nested target
			}
			a.Param, a.Prop = AssertThisProp, prop
		case name == "this" && method:
			if strings.HasPrefix(after, "[") {
				continue
			}
			a.Param = AssertThis
		default:
			if strings.HasPrefix(after, "->") || strings.HasPrefix(after, "[") {
				continue
			}
			for i, p := range ps {
				if p.Var != nil && p.Var.Name == name && !p.Variadic {
					a.Param = i
				}
			}
		}
		if a.Param == -3 {
			continue
		}
		ty := types.FromDoc(typ, x.tplResolver(at))
		if ty.IsUnknown() || ty.Has("mixed") {
			continue
		}
		a.Type = ty.DocString()
		out = append(out, a)
	}
	return out
}

func (x *extractor) funcTemplates(d *phpdoc.Doc, ps []*syntax.Param, asserts []Assertion, at uint32) *FuncTemplates {
	if d == nil || len(x.methodTpl) == 0 {
		return nil
	}
	res := x.tplResolver(at)
	need := false // an assertion type uses a template
	for _, a := range asserts {
		need = need || strings.Contains(a.Type, `\~~`)
	}
	ret := ""
	for _, tag := range []string{"phpstan-return", "psalm-return", "return"} {
		t, ok := d.Tag(tag)
		if !ok {
			continue
		}
		typ, _ := phpdoc.SplitType(t.Text)
		if ds := types.FromDoc(typ, res).DocString(); strings.Contains(ds, `\~~`) {
			ret = ds
			break
		}
	}
	if ret == "" && !need {
		return nil
	}
	ft := &FuncTemplates{Return: ret}
	for _, p := range d.TemplateParams() {
		if x.methodTpl[p.Name] {
			ft.Templates = append(ft.Templates, Template{Name: p.Name, Bound: x.docTypeStr(p.Bound, at)})
		}
	}
	docs := map[string]string{}
	for _, tag := range []string{"param", "psalm-param", "phpstan-param"} { // later wins
		for _, p := range d.ParamsOf(tag) {
			if p.Name != "" && p.Type != "" {
				docs[p.Name] = p.Type
			}
		}
	}
	for i, p := range ps {
		text := docs[p.Var.Name]
		if text == "" || !mentionsAny(text, x.methodTpl) {
			continue
		}
		if ds := types.FromDoc(text, res).DocString(); strings.Contains(ds, `\~~`) {
			if ft.Params == nil {
				ft.Params = make([]string, len(ps))
			}
			ft.Params[i] = ds
		}
	}
	return ft
}

// condReturn returns the conditional return type documented by d
// (preferring @phpstan-return and @psalm-return) in types.Cond canonical
// form, "" when there is none. A template subject (`(T is int ? …)`)
// stands for the parameter documented as exactly T.
func (x *extractor) condReturn(d *phpdoc.Doc, at uint32) string {
	subject := func(name string) string {
		for _, tag := range []string{"phpstan-param", "psalm-param", "param"} {
			for _, p := range d.ParamsOf(tag) {
				if p.Type == name && p.Name != "" {
					return p.Name
				}
			}
		}
		return ""
	}
	for _, tag := range []string{"phpstan-return", "psalm-return", "return"} {
		t, ok := d.Tag(tag)
		if !ok {
			continue
		}
		typ, _ := phpdoc.SplitType(t.Text)
		if !strings.HasPrefix(typ, "(") || !strings.Contains(typ, " is ") {
			continue
		}
		if c, ok := types.ParseCond(typ, x.resolver(at), subject); ok {
			if s := c.String(); len(s) <= types.MaxDocTypeLen {
				return s
			}
		}
	}
	return ""
}

// mentionsAny reports whether doc type text contains one of names as a word
// (a cheap filter before parsing).
func mentionsAny(text string, names map[string]bool) bool {
	start := -1
	for i := 0; i <= len(text); i++ {
		word := i < len(text) && (text[i] == '_' || text[i] == '\\' || text[i] >= 'a' && text[i] <= 'z' ||
			text[i] >= 'A' && text[i] <= 'Z' || text[i] >= '0' && text[i] <= '9' || text[i] >= 0x80)
		if word && start < 0 {
			start = i
		}
		if !word && start >= 0 {
			if names[text[start:i]] {
				return true
			}
			start = -1
		}
	}
	return false
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
	meth.Return, meth.RetVer = x.returnType(m.ReturnType, m.Attrs, at)
	if d != nil {
		meth.DocReturn = x.docTypeStr(d.ReturnType(), at)
		meth.GenReturn = x.genReturn(d, at)
		meth.CondReturn = x.condReturn(d, at)
		meth.Asserts = x.assertions(d, m.Params, !meth.Static, at)
		meth.Tpl = x.funcTemplates(d, m.Params, meth.Asserts, at)
		meth.Deprecated = d.Has("deprecated")
	}
	c.Methods[strings.ToLower(meth.Name)] = meth
	if strings.EqualFold(meth.Name, "__construct") {
		for _, p := range m.Params {
			if len(p.Modifiers) == 0 {
				continue
			}
			c.Props[p.Var.Name] = &Property{Name: p.Var.Name, Class: c.FQN, Visibility: visibility(p.Modifiers),
				Readonly: p.Modifiers.Has(syntax.TReadonly) || c.Readonly, Type: x.typeStr(p.Type, at), Promoted: true,
				HasDefault: p.Default != nil, Default: x.text(p.Default), Span: p.Span()}
		}
	}
}

func (x *extractor) function(n *syntax.Function) {
	d := x.doc(n)
	saved, savedClass := x.methodTpl, x.classTpl
	x.methodTpl, x.classTpl = nil, nil
	if d != nil {
		for _, t := range d.Templates() {
			if x.methodTpl == nil {
				x.methodTpl = map[string]bool{}
			}
			x.methodTpl[t] = true
		}
	}
	x.withTemplates(d, func() { x.functionBody(n, d) })
	x.methodTpl, x.classTpl = saved, savedClass
}

func (x *extractor) functionBody(n *syntax.Function, d *phpdoc.Doc) {
	at := n.Span().Start
	ns := x.r.Namespace(at)
	fqn := n.Name.Value
	if ns != "" {
		fqn = ns + `\` + fqn
	}
	fn := &Function{FQN: fqn, Params: x.params(n.Params, d, at), ByRef: n.ByRef,
		File: x.f.Path, Span: n.Span(), Avail: x.avail(n.Attrs, d)}
	fn.Return, fn.RetVer = x.returnType(n.ReturnType, n.Attrs, at)
	if d != nil {
		fn.DocReturn = x.docTypeStr(d.ReturnType(), at)
		fn.CondReturn = x.condReturn(d, at)
		fn.Asserts = x.assertions(d, n.Params, false, at)
		fn.Tpl = x.funcTemplates(d, n.Params, fn.Asserts, at)
		fn.Deprecated = d.Has("deprecated")
	}
	x.out.Functions = append(x.out.Functions, fn)
}

// returnType uses the declared type, else a stub LanguageLevelTypeAware
// return type: the newest version's type, with the older versions' types
// in byVer (see VerType; nil when the type does not vary).
func (x *extractor) returnType(n syntax.Expr, attrs []*syntax.AttributeGroup, at uint32) (ret string, byVer []VerType) {
	if t := x.typeStr(n, at); t != "" {
		return t, nil
	}
	for _, g := range attrs {
		for _, a := range g.Attrs {
			if !strings.HasSuffix(a.Name.Value, "LanguageLevelTypeAware") || a.Args == nil {
				continue
			}
			def, hasDef := "", false
			var vers []VerType // version map entries: Until = the version it starts at
			for _, arg := range a.Args.Args {
				arg, ok := arg.(*syntax.Arg)
				if !ok {
					continue
				}
				if arg.Name != nil && arg.Name.Value == "default" {
					def, hasDef = x.docTypeStr(unquote(x.text(arg.Value)), at), true
					continue
				}
				arr, ok := arg.Value.(*syntax.Array)
				if !ok {
					continue
				}
				for _, it := range arr.Items {
					if it == nil || it.Value == nil || it.Key == nil {
						continue
					}
					v, err := phpver.Parse(unquote(x.text(it.Key)))
					if err != nil {
						continue
					}
					vers = append(vers, VerType{Until: v, Type: x.docTypeStr(unquote(x.text(it.Value)), at)})
				}
			}
			if len(vers) == 0 {
				if def != "" {
					return def, nil
				}
				continue
			}
			sort.SliceStable(vers, func(i, j int) bool { return vers[i].Until < vers[j].Until })
			// Before the first listed version the default applies; from
			// each listed version on, its type (until the next one).
			prev, prevSet := def, hasDef
			for _, v := range vers {
				if prevSet {
					byVer = append(byVer, VerType{Until: v.Until, Type: prev})
				}
				prev, prevSet = v.Type, true
			}
			if len(byVer) > 0 && byVer[len(byVer)-1].Type == prev && allSame(byVer) {
				byVer = nil // the same type at every version
			}
			return prev, byVer
		}
	}
	return "", nil
}

// allSame reports whether every entry of vs has the same type.
func allSame(vs []VerType) bool {
	for _, v := range vs {
		if v.Type != vs[0].Type {
			return false
		}
	}
	return true
}

func unquote(s string) string {
	if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') {
		return s[1 : len(s)-1]
	}
	return s
}

// avail reads stub availability: #[PhpStormStubsElementAvailable(from:, to:)]
// and @since / @removed doc tags.
func (x *extractor) avail(attrs []*syntax.AttributeGroup, d *phpdoc.Doc) Avail {
	var a Avail
	for _, g := range attrs {
		for _, at := range g.Attrs {
			if !strings.HasSuffix(at.Name.Value, "PhpStormStubsElementAvailable") || at.Args == nil {
				continue
			}
			for i, arg := range at.Args.Args {
				arg, ok := arg.(*syntax.Arg)
				if !ok {
					continue
				}
				v, err := phpver.Parse(unquote(x.text(arg.Value)))
				if err != nil {
					continue
				}
				name := ""
				if arg.Name != nil {
					name = arg.Name.Value
				}
				switch {
				case name == "from" || (name == "" && i == 0):
					a.From = v
				case name == "to" || (name == "" && i == 1):
					a.To = v
				}
			}
		}
	}
	if d != nil {
		if t, ok := d.Tag("since"); ok && a.From == 0 {
			if v, err := phpver.Parse(firstWord(t.Text)); err == nil {
				a.From = v
			}
		}
		if t, ok := d.Tag("removed"); ok && a.To == 0 {
			if v, err := phpver.Parse(firstWord(t.Text)); err == nil && v > phpver.Min {
				a.To = v - 1
			}
		}
	}
	return a
}

func firstWord(s string) string {
	if i := strings.IndexAny(s, " \t\n"); i >= 0 {
		return s[:i]
	}
	return s
}

// define() calls with a literal name declare global constants.
func (x *extractor) define(call *syntax.FuncCall) {
	name, ok := call.Name.(*syntax.Name)
	if !ok || !strings.EqualFold(strings.TrimPrefix(name.Value, `\`), "define") || call.Args == nil || len(call.Args.Args) < 2 {
		return
	}
	a0, ok0 := call.Args.Args[0].(*syntax.Arg)
	a1, ok1 := call.Args.Args[1].(*syntax.Arg)
	if !ok0 || !ok1 {
		return
	}
	lit, ok := a0.Value.(*syntax.Literal)
	if !ok || lit.LitKind != syntax.LitString {
		return
	}
	x.out.Constants = append(x.out.Constants, &Constant{FQN: strings.TrimPrefix(unquote(lit.Raw), `\`), Value: x.text(a1.Value), File: x.f.Path, Span: call.Span()})
}
