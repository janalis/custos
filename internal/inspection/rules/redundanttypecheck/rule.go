// Package redundanttypecheck implements the native RedundantTypeCheck
// inspection.
package redundanttypecheck

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "RedundantTypeCheck" }

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KBinary, syntax.KInstanceof, syntax.KFuncCall}
}

func (rule) Semantic() {}

// Check reports a type test whose answer the proven type of the tested
// value decides.
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if c, ok := semanticquery.DecidedTypeCheck(ctx, n); ok && !c.Comparison {
		ctx.ReportNode(c.Node, c.Message)
	}
}
