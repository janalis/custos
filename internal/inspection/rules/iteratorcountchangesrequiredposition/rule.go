// Package iteratorcountchangesrequiredposition implements IteratorCountChangesRequiredPosition.
package iteratorcountchangesrequiredposition

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Restore the iterator position after counting."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "IteratorCountChangesRequiredPosition" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckIteratorCountChangesRequiredPosition(ctx, n, message)
}
