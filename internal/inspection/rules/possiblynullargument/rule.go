// Package possiblynullargument implements the PossiblyNullArgument inspection.
package possiblynullargument

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "PossiblyNullArgument" }

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall, syntax.KMethodCall, syntax.KStaticCall, syntax.KNew}
}

func (rule) Semantic() {}

// Check reports an argument native declarations prove may be null, passed
// to a parameter whose native type rejects null.
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	for _, a := range semanticquery.PossiblyNullArguments(ctx, n) {
		ctx.ReportNode(a.Arg.Value, a.Message())
	}
}
