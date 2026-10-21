// Package pdofetchcolumnfalsyvalueloss implements the native PdoFetchColumnFalsyValueLoss inspection.
package pdofetchcolumnfalsyvalueloss

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare fetch failure strictly with false."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoFetchColumnFalsyValueLoss" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)

	if !semanticquery.NativeMethod(ctx, c, "PDOStatement", "fetchColumn") {
		return
	}
	assignment, ok := c.Parent().(*syntax.Assign)
	if !ok {
		return
	}
	p := assignment.Parent()
	for {
		par, ok := p.(*syntax.Paren)
		if !ok {
			break
		}
		p = par.Parent()
	}
	if branch, ok := p.(*syntax.If); ok && branch.Cond != nil {
		ctx.ReportNode(assignment, message)
	}
}
