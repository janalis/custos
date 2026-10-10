// Package domcrossdocumentappend implements DomCrossDocumentAppend.
package domcrossdocumentappend

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Import the node into the destination document."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DomCrossDocumentAppend" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "DOMNode", "appendChild") && !semanticquery.NativeMethod(ctx, c, "DOMNode", "insertBefore") && !semanticquery.NativeMethod(ctx, c, "DOMNode", "replaceChild") {
		return
	}
	a := semanticquery.ExpansionDOMOwner(ctx, c.Var)
	b := semanticquery.ExpansionDOMOwner(ctx, semanticquery.CallArgument(c.Args, 0, "node"))
	if a != nil && b != nil && a != b {
		ctx.ReportNode(c, message)
	}
}
