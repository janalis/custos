// Package curlreadcallbackexceedsrequestedsize implements CurlReadCallbackExceedsRequestedSize.
package curlreadcallbackexceedsrequestedsize

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return no more than the requested byte count."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlReadCallbackExceedsRequestedSize" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckCurlReadCallbackExceedsRequestedSize(ctx, n, message)
}
