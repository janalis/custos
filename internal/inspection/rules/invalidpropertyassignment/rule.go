// Package invalidpropertyassignment implements the native
// InvalidPropertyAssignment inspection.
package invalidpropertyassignment

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "InvalidPropertyAssignment" }

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KAssign, syntax.KForeach, syntax.KIncDec}
}

func (rule) Semantic() {}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	for _, m := range semanticquery.PropertyWriteMismatches(ctx, n) {
		if m.Proven {
			ctx.ReportNode(m.Value, m.Message())
		}
	}
}
