// Package fiberreturnbeforetermination implements FiberReturnBeforeTermination.
package fiberreturnbeforetermination

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Read the fiber return after termination."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FiberReturnBeforeTermination" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckFiberReturnBeforeTermination(ctx, n, message)
}
