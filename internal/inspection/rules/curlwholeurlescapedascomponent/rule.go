// Package curlwholeurlescapedascomponent implements CurlWholeUrlEscapedAsComponent.
package curlwholeurlescapedascomponent

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Escape URL components before assembling the URL."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlWholeUrlEscapedAsComponent" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckCurlWholeURLEscapedAsComponent(ctx, n, message)
}
