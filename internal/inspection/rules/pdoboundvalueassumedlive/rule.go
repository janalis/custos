// Package pdoboundvalueassumedlive implements PdoBoundValueAssumedLive.
package pdoboundvalueassumedlive

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Rebind the changed value before execution."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoBoundValueAssumedLive" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPdoBoundValueAssumedLive(ctx, n, message)
}
