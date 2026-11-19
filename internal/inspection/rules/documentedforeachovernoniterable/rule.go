// Package documentedforeachovernoniterable implements the native
// DocumentedForeachOverNonIterable inspection.
package documentedforeachovernoniterable

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "DocumentedForeachOverNonIterable" }

func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KForeach} }

func (rule) Semantic() {}

// Check reports the subject of a loop when PHPDoc-aware types show it holds
// only scalars or null and native declarations alone do not (those are
// ForeachOverNonIterable's).
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	loop := n.(*syntax.Foreach)
	if j, ok := semanticquery.NonIterableSubject(ctx, loop); ok && !j.Proven {
		ctx.ReportNode(loop.Expr, "Iterate over an array or an object, not "+j.Type.String()+".")
	}
}
