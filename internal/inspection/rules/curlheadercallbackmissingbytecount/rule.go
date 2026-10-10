// Package curlheadercallbackmissingbytecount implements CurlHeaderCallbackMissingByteCount.
package curlheadercallbackmissingbytecount

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return the consumed header byte count."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlHeaderCallbackMissingByteCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	semanticquery.CheckCurlHeaderCallbackMissingByteCount(ctx, n, message)
}
