// Package transferencodingwithcontentlength implements the native TransferEncodingWithContentLength inspection.
package transferencodingwithcontentlength

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Choose a single HTTP message framing mechanism."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "TransferEncodingWithContentLength" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	key, _, _, known := semanticquery.ExpansionHeader(ctx, c)
	if !known || (key != "content-length" && key != "transfer-encoding") {
		return
	}
	response := semanticquery.ExpansionResponseAt(ctx, c)
	if response.HasLength && response.HasTransfer {
		ctx.ReportNode(c, message)
	}
}
