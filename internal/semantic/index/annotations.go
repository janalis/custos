package index

import (
	"sort"
	"strings"

	"custos/internal/php/phpdoc"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/types"
)

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
		c.Templates = append(c.Templates, Template{Name: p.Name, Bound: x.docTypeStr(p.Bound, at), Default: x.docTypeStr(p.Default, at)})
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
	if ds := types.FromDoc(d.EffectiveReturnType(), x.genResolver(at)).DocString(); strings.Contains(ds, `\~`) {
		return ds
	}
	return ""
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
		sig, ok := phpdoc.ParseMethod(t.Text)
		if !ok {
			continue
		}
		key := strings.ToLower(sig.Name)
		if _, exists := c.Methods[key]; !exists {
			var params []Param
			for _, p := range sig.Params {
				params = append(params, Param{
					Name: p.Name, DocType: x.docTypeStr(p.Type, at), Default: p.Default,
					Optional: p.Optional, ByRef: p.ByRef, Variadic: p.Variadic,
				})
			}
			c.Methods[key] = &Method{
				Name: sig.Name, Class: c.FQN, Static: sig.Static,
				DocReturn: x.docTypeStr(sig.Return, at), Params: params, Magic: true,
			}
		}
	}
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
	if ds := types.FromDoc(d.EffectiveReturnType(), res).DocString(); strings.Contains(ds, `\~~`) {
		ret = ds
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
	for _, p := range d.EffectiveParams() {
		docs[p.Name] = p.Type
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
		for _, p := range d.EffectiveParams() {
			if p.Type == name && p.Name != "" {
				return p.Name
			}
		}
		return ""
	}
	typ := d.EffectiveReturnType()
	if strings.HasPrefix(typ, "(") && strings.Contains(typ, " is ") {
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
					v, err := phpversion.Parse(unquote(x.text(it.Key)))
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
				v, err := phpversion.Parse(unquote(x.text(arg.Value)))
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
			if v, err := phpversion.Parse(firstWord(t.Text)); err == nil {
				a.From = v
			}
		}
		if t, ok := d.Tag("removed"); ok && a.To == 0 {
			if v, err := phpversion.Parse(firstWord(t.Text)); err == nil && v > phpversion.Min {
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
