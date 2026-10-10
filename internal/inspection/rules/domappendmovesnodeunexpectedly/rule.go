// Package domappendmovesnodeunexpectedly implements DomAppendMovesNodeUnexpectedly.
package domappendmovesnodeunexpectedly

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Clone the node when preserving its existing parent."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DomAppendMovesNodeUnexpectedly" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "DOMNode", "appendChild") {
		return
	}
	node := semanticquery.CallArgument(c.Args, 0, "node")
	var previous *syntax.MethodCall
	for _, record := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
		p, ok := record.Node.(*syntax.MethodCall)
		if !ok || !semanticquery.NativeDominates(p, c) {
			continue
		}
		if !semanticquery.ExpansionSameObject(ctx, semanticquery.CallArgument(p.Args, 0, "node"), node) {
			continue
		}

		if semanticquery.NativeMethod(ctx, p, "DOMNode", "removeChild") || semanticquery.NativeMethod(ctx, p, "DOMNode", "replaceChild") {
			previous = nil
			continue
		}
		if semanticquery.NativeMethod(ctx, p, "DOMNode", "appendChild") {
			previous = p
		}

	}
	if previous != nil {
		old := semanticquery.ExpansionDOMOwner(ctx, previous.Var)
		current := semanticquery.ExpansionDOMOwner(ctx, c.Var)
		if old != nil && current != nil && !semanticquery.ExpansionSameObject(ctx, previous.Var, c.Var) {
			ctx.ReportNode(c, message)
		}

	}
}
