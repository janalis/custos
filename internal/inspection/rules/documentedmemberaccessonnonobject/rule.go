// Package documentedmemberaccessonnonobject implements the native
// DocumentedMemberAccessOnNonObject inspection.
package documentedmemberaccessonnonobject

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }

type rule struct{}

func (rule) ID() string { return "DocumentedMemberAccessOnNonObject" }

func (rule) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KPropertyFetch, syntax.KMethodCall}
}

func (rule) Semantic() {}

// Check reports the receiver of a member access when PHPDoc-aware types
// show it is never an object and native declarations alone do not (those
// are MemberAccessOnNonObject's).
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if receiver, j, ok := semanticquery.NonObjectReceiver(ctx, n); ok && !j.Proven {
		ctx.ReportNode(receiver, "Access this member on an object, not "+j.Type.String()+".")
	}
}
