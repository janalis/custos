// Package nonrepeatableattributerepeated implements the native NonRepeatableAttributeRepeated inspection.
package nonrepeatableattributerepeated

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Remove the repeated attribute."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "NonRepeatableAttributeRepeated" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAttribute} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	a := n.(*syntax.Attribute)
	if ctx.PHP < phpversion.PHP80 {
		return
	}
	fqn, flags, known := semanticquery.NativeAttributeFlags(ctx, a)
	repeat := int64(64)
	if ctx.PHP >= phpversion.PHP85 {
		repeat = 128
	}
	if !known || flags&repeat != 0 {
		return
	}
	owner := a.Parent().Parent()
	found := false
	syntax.Children(owner, func(n syntax.Node) {
		g, ok := n.(*syntax.AttributeGroup)
		if !ok {
			return
		}
		for _, prev := range g.Attrs {
			if prev.Span().Start < a.Span().Start && strings.EqualFold(ctx.Names().Class(prev.Name.Value, prev.Span().Start), fqn) {
				found = true
			}
		}
	})
	if found {
		ctx.ReportNode(a, message)
	}
}
