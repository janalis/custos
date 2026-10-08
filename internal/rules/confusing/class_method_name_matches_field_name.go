package confusing

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/phpdoc"
	"custos/internal/syntax"
	"custos/internal/types"
)

// classMethodNameMatchesFieldName reports methods sharing their name with a
// property whose type is unknown or callable.
type classMethodNameMatchesFieldName struct{}

func init() { register(classMethodNameMatchesFieldName{}) }

func (classMethodNameMatchesFieldName) ID() string { return "ClassMethodNameMatchesFieldName" }

func (classMethodNameMatchesFieldName) Semantic() {}

func (classMethodNameMatchesFieldName) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KMethod}
}

const (
	cmnUnknownMsg  = "A property with this name exists and its type is unknown; rename the method or type the property."
	cmnCallableMsg = "A callable property with this name exists; rename the method (for example with a get/is/has prefix)."
)

func (r classMethodNameMatchesFieldName) Check(ctx *analysis.Context, n syntax.Node) {
	m := n.(*syntax.Method)
	if m.Name == nil || m.Name.Span().Len() == 0 {
		return
	}
	cls, ok := m.Parent().(*syntax.ClassLike)
	if !ok || cls.ClassKind == syntax.KindInterface { // D1 / E1
		return
	}
	fqn := ctx.Types().ClassFQN(cls)
	var prop *index.Property
	if fqn != "" {
		prop = ctx.Index().FindProperty(fqn, m.Name.Value, ctx.PHP)
	} else { // anonymous class: own members only
		prop = localProperty(ctx, cls, m.Name.Value)
	}
	if prop == nil || prop.Magic { // D2
		return
	}
	atoms := r.propertyTypes(ctx, prop)
	if len(atoms) == 0 { // D4
		ctx.Report(m.Name.Span(), cmnUnknownMsg)
		return
	}
	for _, a := range atoms { // D5
		low := strings.ToLower(a)
		if low == "callable" || low == `\closure` || low == "closure" {
			ctx.Report(m.Name.Span(), cmnCallableMsg)
			return
		}
	}
}

func localProperty(ctx *analysis.Context, cls *syntax.ClassLike, name string) *index.Property {
	for _, mem := range cls.Members {
		if p, ok := mem.(*syntax.Property); ok {
			for _, it := range p.Props {
				if it.Var.Name == name {
					return &index.Property{Name: name, Span: it.Span(), Default: ctx.Text(it.Default), HasDefault: it.Default != nil}
				}
			}
		}
	}
	return nil
}

// propertyTypes unions the declared type, the preceding @var comment and the
// default value's type (D3).
func (classMethodNameMatchesFieldName) propertyTypes(ctx *analysis.Context, p *index.Property) []string {
	var all []string
	add := func(t types.Type) {
		if !t.IsUnknown() {
			all = append(all, t.Atoms()...)
		}
	}
	item, decl := findPropertyItem(ctx.File, p.Span)
	if item != nil {
		at := decl.Span().Start
		res := func(w string) string { return ctx.Names().Class(w, at) }
		add(types.FromNode(decl.Type, res))
		if c := precedingComment(ctx.File, decl); c != "" {
			if strings.HasPrefix(c, "/*") && !strings.HasPrefix(c, "/**") {
				c = "/**" + c[2:]
			}
			add(types.FromDoc(phpdoc.Parse(c).VarType(p.Name), res))
		}
		if item.Default != nil {
			add(ctx.TypeOf(item.Default))
		}
	} else {
		add(types.FromDoc(p.Type, nil))
		add(types.FromDoc(p.DocType, nil))
		if p.Promoted { // custos: the constructor's @param documents it
			if c := ctx.Index().Class(p.Class, ctx.PHP); c != nil {
				if ctor := c.Methods["__construct"]; ctor != nil {
					for _, prm := range ctor.Params {
						if prm.Name == p.Name {
							add(types.FromDoc(prm.DocType, nil))
						}
					}
				}
			}
		}
		if p.HasDefault {
			add(infer.LiteralTextType(p.Default))
		}
	}
	return all
}

// findPropertyItem locates the property item declared at span in f.
func findPropertyItem(f *syntax.File, span syntax.Span) (*syntax.PropertyItem, *syntax.Property) {
	var item *syntax.PropertyItem
	var decl *syntax.Property
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if item != nil {
			return false
		}
		if it, ok := n.(*syntax.PropertyItem); ok && it.Span() == span {
			if d, ok := it.Parent().(*syntax.Property); ok {
				item, decl = it, d
			}
			return false
		}
		return true
	})
	return item, decl
}

// precedingComment returns the comment (doc or block) right before n,
// skipping whitespace only.
func precedingComment(f *syntax.File, n syntax.Node) string {
	start := n.Span().Start
	toks := f.Tokens
	lo, hi := 0, len(toks)
	for lo < hi {
		mid := (lo + hi) / 2
		if toks[mid].Start < start {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	for j := lo - 1; j >= 0; j-- {
		t := toks[j]
		switch t.Kind {
		case syntax.TWhitespace:
			continue
		case syntax.TDocComment, syntax.TComment:
			s := string(f.Src[t.Start:t.End])
			if strings.HasPrefix(s, "/*") {
				return s
			}
		}
		break
	}
	return ""
}
