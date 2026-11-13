// Package documentedargumenttype implements the DocumentedArgumentType
// inspection.
package documentedargumenttype

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
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
