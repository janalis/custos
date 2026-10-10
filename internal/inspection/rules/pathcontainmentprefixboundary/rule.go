// Package pathcontainmentprefixboundary implements the native PathContainmentPrefixBoundary inspection.
package pathcontainmentprefixboundary

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Require a directory boundary in the containment check."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PathContainmentPrefixBoundary" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPathContainmentPrefixBoundary(ctx, n, message)
}
