// Package domtextassignmentdestroyschildren implements DomTextAssignmentDestroysChildren.
package domtextassignmentdestroyschildren

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Preserve child nodes when adding element text."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DomTextAssignmentDestroysChildren" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	a := n.(*syntax.Assign)
	p, ok := a.Var.(*syntax.PropertyFetch)
	if !ok {
		return
	}
	id, ok := p.Name.(*syntax.Identifier)
	if !ok || (id.Value != "nodeValue" && id.Value != "textContent") {
		return
	}
	if semanticquery.ExpansionDOMElementChildren(ctx, p.Var, a) || semanticquery.ExpansionDOMBeforeWrite(ctx, p.Var, a) || semanticquery.ExpansionDOMCreatedBeforeWrite(ctx, p.Var, a) {
		ctx.ReportNode(a, message)
	}
}
