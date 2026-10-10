// Package overrideattributewithouttarget implements the native OverrideAttributeWithoutTarget inspection.
package overrideattributewithouttarget

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
)

const message = "Apply Override to an inherited member."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "OverrideAttributeWithoutTarget" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAttribute} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	a := n.(*syntax.Attribute)
	if ctx.PHP < phpversion.PHP83 {
		return
	}
	name := ctx.Names().Class(a.Name.Value, a.Name.Span().Start)
	bc := ctx.Index().Class(name, ctx.PHP)
	if !strings.EqualFold(name, "Override") || bc == nil || !strings.HasPrefix(bc.File, "stubs/") {
		return
	}
	owner := a.Parent().Parent()
	decl := syntax.EnclosingClass(owner)
	if decl == nil {
		return
	}
	fqn := ctx.Names().DeclFQN(decl)
	c := ctx.Index().Class(fqn, ctx.PHP)
	if c == nil || !ctx.Index().AncestorsComplete(fqn, ctx.PHP) {
		return
	}
	member := ""
	property := false
	switch x := owner.(type) {
	case *syntax.Method:
		member = x.Name.Value
	case *syntax.Property:
		if ctx.PHP < phpversion.PHP85 || len(x.Props) != 1 {
			return
		}
		member = x.Props[0].Var.Name
		property = true
	default:
		return
	}
	for _, anc := range ctx.Index().Ancestors(fqn, ctx.PHP) {
		if anc.Parent != "" && ctx.Index().Class(anc.Parent, ctx.PHP) == nil {
			return
		}
		for _, super := range anc.Interfaces {
			if ctx.Index().Class(super, ctx.PHP) == nil {
				return
			}
		}
		if strings.EqualFold(anc.FQN, fqn) {
			continue
		}
		if len(anc.Traits) > 0 {
			return
		}
		if property {
			p := ctx.Index().FindProperty(anc.FQN, member, ctx.PHP)
			if p != nil && p.Visibility != index.Private {
				return
			}
		} else {
			m := ctx.Index().FindMethod(anc.FQN, member, ctx.PHP)
			if m != nil && m.Visibility != index.Private {
				return
			}
		}
	}
	if len(c.Traits) > 0 {
		return
	}
	ctx.ReportNode(a.Parent(), message)
}
