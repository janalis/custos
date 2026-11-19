// Package documentedargumenttype implements the native
// DocumentedArgumentType inspection.
package documentedargumenttype

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "DocumentedArgumentType" }

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall, syntax.KStaticCall, syntax.KNew}
}

func (rule) Semantic() {}

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	for _, m := range semanticquery.ArgumentTypeMismatches(ctx, n) {
		if !m.Proven {
			ctx.ReportNode(m.Arg.Value, m.Message())
		}
	}
}
