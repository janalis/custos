// Package conflictingcontentlengthheaders implements the native ConflictingContentLengthHeaders inspection.
package conflictingcontentlengthheaders

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Send one consistent Content-Length value."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ConflictingContentLengthHeaders" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	key, value, replace, known := semanticquery.ExpansionHeader(ctx, c)
	if !known || key != "content-length" || replace {
		return
	}
	size, known := semanticquery.ExpansionDecimalLength(value)
	if !known {
		return
	}
	response := semanticquery.ExpansionResponseAt(ctx, c)
	for _, previous := range response.PreviousLengths {
		n, ok := semanticquery.ExpansionDecimalLength(previous)
		if ok && n != size {
			ctx.ReportNode(c, message)
			return
		}
	}
}
