// Package curlheaderoptionrequiresarray implements CurlHeaderOptionRequiresArray.
package curlheaderoptionrequiresarray

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Supply headers as an array."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlHeaderOptionRequiresArray" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckCurlHeaderOptionRequiresArray(ctx, n, message)
}
