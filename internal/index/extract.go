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
		if x.templates[w] {
			return ""
		}
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
	if d != nil {
		for _, p := range d.Params() {
			docTypes[p.Name] = p.Type
		}
	}
	out := make([]Param, 0, len(ps))
	for _, p := range ps {
		out = append(out, Param{
			Name: p.Var.Name, Type: x.typeStr(p.Type, at), DocType: x.docTypeStr(docTypes[p.Var.Name], at),
			Optional: p.Default != nil || p.Variadic, ByRef: p.ByRef, Variadic: p.Variadic,
			Promoted: len(p.Modifiers) > 0, Default: x.text(p.Default), Avail: x.avail(p.Attrs, nil),
		})
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
	x.withTemplates(d, func() { x.methodBody(c, m) })
	x.methodTpl = saved
}

func (x *extractor) methodBody(c *Class, m *syntax.Method) {
	at := m.Span().Start
	d := x.doc(m)
	meth := &Method{
		Name: m.Name.Value, Class: c.FQN, Visibility: visibility(m.Modifiers), Static: m.Modifiers.Has(syntax.TStatic),
		Abstract: m.Modifiers.Has(syntax.TAbstract) || c.Kind == syntax.KindInterface, Final: m.Modifiers.Has(syntax.TFinal),
		ByRef: m.ByRef, Params: x.params(m.Params, d, at), Return: x.returnType(m.ReturnType, m.Attrs, at), Span: m.Span(),
		Avail: x.avail(m.Attrs, d),
	}
	if d != nil {
		meth.DocReturn = x.docTypeStr(d.ReturnType(), at)
		meth.GenReturn = x.genReturn(d, at)
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
	x.withTemplates(x.doc(n), func() { x.functionBody(n) })
}

func (x *extractor) functionBody(n *syntax.Function) {
	at := n.Span().Start
	ns := x.r.Namespace(at)
	fqn := n.Name.Value
	if ns != "" {
		fqn = ns + `\` + fqn
	}
	d := x.doc(n)
	fn := &Function{FQN: fqn, Params: x.params(n.Params, d, at), Return: x.returnType(n.ReturnType, n.Attrs, at), ByRef: n.ByRef,
		File: x.f.Path, Span: n.Span(), Avail: x.avail(n.Attrs, d)}
	if d != nil {
		fn.DocReturn = x.docTypeStr(d.ReturnType(), at)
		fn.Deprecated = d.Has("deprecated")
	}
	x.out.Functions = append(x.out.Functions, fn)
}

// returnType uses the declared type, else a stub LanguageLevelTypeAware default.
func (x *extractor) returnType(n syntax.Expr, attrs []*syntax.AttributeGroup, at uint32) string {
	if t := x.typeStr(n, at); t != "" {
		return t
	}
	for _, g := range attrs {
		for _, a := range g.Attrs {
			if !strings.HasSuffix(a.Name.Value, "LanguageLevelTypeAware") || a.Args == nil {
				continue
			}
			best := ""
			for _, arg := range a.Args.Args {
				arg, ok := arg.(*syntax.Arg)
				if !ok {
					continue
				}
				if arg.Name != nil && arg.Name.Value == "default" {
					if best == "" {
						best = unquote(x.text(arg.Value))
					}
					continue
				}
				// Version map: take the highest version's type.
				if arr, ok := arg.Value.(*syntax.Array); ok && len(arr.Items) > 0 {
					if last := arr.Items[len(arr.Items)-1]; last != nil && last.Value != nil {
						best = unquote(x.text(last.Value))
					}
				}
			}
			if best != "" {
				return x.docTypeStr(best, at)
			}
		}
	}
	return ""
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
