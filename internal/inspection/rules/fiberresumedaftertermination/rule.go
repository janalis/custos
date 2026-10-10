// Package fiberresumedaftertermination implements FiberResumedAfterTermination.
package fiberresumedaftertermination

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Resume a fiber before it terminates."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "FiberResumedAfterTermination" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckFiberResumedAfterTermination(ctx, n, message)
}
