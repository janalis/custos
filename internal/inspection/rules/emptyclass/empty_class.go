package emptyclass

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// emptyClass reports named classes, traits and enums declaring nothing.
type emptyClass struct{}

func (emptyClass) ID() string               { return "EmptyClass" }
func (emptyClass) Semantic()                {}
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
	if semanticquery.DocHasTag(ctx.File, cl, "deprecated") { // D3
		return
	}
	// custos: an attribute class is used by its name alone, and any other
	// attribute configures the class from outside (#[ApiResource],
	// #[ORM\Entity], #[Route], #[AsEventListener]): declaring nothing is
	// its purpose.
	for _, g := range cl.Attrs {
		if len(g.Attrs) > 0 {
			return
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
