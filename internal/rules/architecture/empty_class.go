package architecture

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// emptyClass reports named classes, traits and enums declaring nothing.
type emptyClass struct{}

func init() { register(emptyClass{}) }

func (emptyClass) ID() string { return "EmptyClass" }

func (emptyClass) Semantic() {}

func (emptyClass) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KClassLike} }

func (emptyClass) Check(ctx *analysis.Context, n syntax.Node) {
	cl := n.(*syntax.ClassLike)
	if cl.Name == nil || cl.ClassKind == syntax.KindInterface || cl.Name.Span().Len() == 0 { // D1
		return
	}
	for _, m := range cl.Members { // D2
		switch m := m.(type) {
		case *syntax.Property, *syntax.ClassConst, *syntax.Method, *syntax.TraitUse, *syntax.EnumCase:
			if m.Span().Len() > 0 {
				return
			}
		}
	}
	if util.DocHasTag(ctx.File, cl, "deprecated") { // D3
		return
	}
	for _, g := range cl.Attrs { // E: an attribute class is used by its name alone
		for _, a := range g.Attrs {
			if a.Name != nil && strings.EqualFold(strings.TrimPrefix(ctx.Names().Class(a.Name.Value, a.Name.Span().Start), `\`), "Attribute") {
				return
			}
		}
	}
	if parent := ctx.Names().ParentFQN(cl); parent != "" { // D4
		ix := ctx.Index()
		if pc := ix.Class(parent, ctx.PHP); pc != nil {
			if pc.Abstract {
				return
			}
			for _, a := range ix.Ancestors(ctx.Names().DeclFQN(cl), ctx.PHP) {
				if strings.TrimPrefix(a.FQN, `\`) == "Exception" {
					return
				}
			}
		}
	}
	msg := "This class declares no members; remove it or give it a purpose."
	if cl.ClassKind == syntax.KindEnum {
		msg = "This enum declares no cases or methods."
	}
	ctx.ReportNode(cl.Name, msg)
}
