// Package memberaccessonnonobject implements the native
// MemberAccessOnNonObject inspection.
package memberaccessonnonobject

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "MemberAccessOnNonObject" }

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KPropertyFetch, syntax.KMethodCall}
}

func (rule) Semantic() {}

// Check reports the receiver of a member access when native declarations prove it is never an object.
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if receiver, j, ok := semanticquery.NonObjectReceiver(ctx, n); ok && j.Proven {
		ctx.ReportNode(receiver, "Access this member on an object, not "+j.Type.String()+".")
	}
}
