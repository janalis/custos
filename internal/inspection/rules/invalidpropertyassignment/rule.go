// Package invalidpropertyassignment implements the InvalidPropertyAssignment
// inspection.
package invalidpropertyassignment

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "InvalidPropertyAssignment" }

func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }

func (rule) Semantic() {}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if m, ok := semanticquery.PropertyAssignmentMismatch(ctx, n); ok && m.Proven {
		ctx.ReportNode(m.Value, m.Message())
	}
}
