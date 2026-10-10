// Package curlresponsezerorejected implements CurlResponseZeroRejected.
package curlresponsezerorejected

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare cURL transport failure strictly with false."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlResponseZeroRejected" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckCurlResponseZeroRejected(ctx, n, message)
}
