// Package pdorepeatednamedmarkerwithoutemulation implements PdoRepeatedNamedMarkerWithoutEmulation.
package pdorepeatednamedmarkerwithoutemulation

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use a distinct marker for each native parameter."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoRepeatedNamedMarkerWithoutEmulation" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckPdoRepeatedNamedMarkerWithoutEmulation(ctx, n, message)
}
