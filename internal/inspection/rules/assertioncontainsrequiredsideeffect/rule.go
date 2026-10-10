// Package assertioncontainsrequiredsideeffect implements AssertionContainsRequiredSideEffect.
package assertioncontainsrequiredsideeffect

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Move required assignments outside assertions."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AssertionContainsRequiredSideEffect" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckAssertionContainsRequiredSideEffect(ctx, n, message)
}
