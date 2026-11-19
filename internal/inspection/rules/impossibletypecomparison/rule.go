// Package impossibletypecomparison implements the native
// ImpossibleTypeComparison inspection.
package impossibletypecomparison

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "ImpossibleTypeComparison" }

func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KBinary} }

func (rule) Semantic() {}

// Check reports an identity comparison between operands whose proven types
// share no value.
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if c, ok := semanticquery.DecidedTypeCheck(ctx, n); ok && c.Comparison {
		ctx.ReportNode(c.Node, c.Message)
	}
}
