// Package contentlengthknownbodymismatch implements the native ContentLengthKnownBodyMismatch inspection.
package contentlengthknownbodymismatch

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match Content-Length to the emitted bytes."

type rule struct{}

// New returns the ContentLengthKnownBodyMismatch inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ContentLengthKnownBodyMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	key, value, _, known := semanticquery.LiteralHeader(ctx, c)
	if !known || key != "content-length" {
		return
	}
	size, known := semanticquery.ParseDecimalLength(value)
	if !known {
		return
	}
	response := semanticquery.ResponseStateAt(ctx, c)
	if !response.Final {
		return
	}
	if response.BodyKnown && response.BodyLength != size {
		ctx.ReportNode(c, message)
	}
}
