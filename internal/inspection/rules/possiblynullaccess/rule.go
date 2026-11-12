// Package possiblynullaccess implements the PossiblyNullAccess inspection.
package possiblynullaccess

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "PossiblyNullAccess" }

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{
		syntax.KPropertyFetch, syntax.KMethodCall, syntax.KStaticCall,
		syntax.KClassConstFetch, syntax.KStaticPropertyFetch, syntax.KArrayDimFetch,
	}
}

func (rule) Semantic() {}

// Check reports the receiver of an access that fails or warns on null when
// native declarations prove the receiver may be null, unless an earlier
// access on the same value already says so.
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if a, ok := semanticquery.PossiblyNullAccess(ctx, n); ok && !a.Repeated {
		ctx.ReportNode(a.Receiver, a.Message())
	}
}
