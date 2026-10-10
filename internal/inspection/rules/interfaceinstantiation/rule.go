// Package interfaceinstantiation implements InterfaceInstantiation.
package interfaceinstantiation

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Instantiate an implementing class."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "InterfaceInstantiation" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNew} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckInterfaceInstantiation(ctx, n, message)
}
