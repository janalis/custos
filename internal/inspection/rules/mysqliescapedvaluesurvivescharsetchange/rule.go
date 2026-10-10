// Package mysqliescapedvaluesurvivescharsetchange implements MysqliEscapedValueSurvivesCharsetChange.
package mysqliescapedvaluesurvivescharsetchange

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Escape values after selecting the connection charset."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MysqliEscapedValueSurvivesCharsetChange" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckMysqliEscapedValueSurvivesCharsetChange(ctx, n, message)
}
