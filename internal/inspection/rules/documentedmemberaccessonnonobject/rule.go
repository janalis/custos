// Package documentedmemberaccessonnonobject implements the native
// DocumentedMemberAccessOnNonObject inspection.
package documentedmemberaccessonnonobject

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
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
	if receiver, t, ok := semanticquery.NonObjectReceiver(ctx, n, true); ok {
		ctx.ReportNode(receiver, "Access this member on an object, not "+t.String()+".")
	}
}
