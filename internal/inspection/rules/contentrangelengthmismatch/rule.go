// Package contentrangelengthmismatch implements the native ContentRangeLengthMismatch inspection.
package contentrangelengthmismatch

import (
	"regexp"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Match the body to the inclusive byte range."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ContentRangeLengthMismatch" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	key, value, _, known := semanticquery.ExpansionHeader(ctx, c)
	if !known || key != "content-range" {
		return
	}
	m := byteRange.FindStringSubmatch(value)
	if len(m) == 0 {
		return
	}
	a, ak := semanticquery.ExpansionDecimalLength(m[1])
	b, bk := semanticquery.ExpansionDecimalLength(m[2])
	total, tk := semanticquery.ExpansionDecimalLength(m[3])
	if !ak || !bk || !tk || a > b || b >= total {
		return
	}
	response := semanticquery.ExpansionResponseAt(ctx, c)
	if response.Final && response.Status == 206 && response.BodyKnown && response.BodyLength != b-a+1 {
		ctx.ReportNode(c, message)
	}
}

var byteRange = regexp.MustCompile(`(?i)^bytes (\d+)-(\d+)/(\d+)$`)
