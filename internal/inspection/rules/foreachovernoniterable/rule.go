// Package foreachovernoniterable implements the native
// ForeachOverNonIterable inspection.
package foreachovernoniterable

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "ForeachOverNonIterable" }

func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KForeach} }

func (rule) Semantic() {}

// Check reports the subject of a loop when native declarations prove it holds only scalars or null.
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	loop := n.(*syntax.Foreach)
	if t, ok := semanticquery.NonIterableSubject(ctx, loop, false); ok {
		ctx.ReportNode(loop.Expr, "Iterate over an array or an object, not "+t.String()+".")
	}
}
