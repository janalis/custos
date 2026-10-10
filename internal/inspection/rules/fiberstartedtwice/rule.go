// Package fiberstartedtwice implements FiberStartedTwice.
package fiberstartedtwice

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Start each fiber only once."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FiberStartedTwice" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckFiberStartedTwice(ctx, n, message)
}
