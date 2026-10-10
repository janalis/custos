// Package bytelengthusedascharactercount implements the ByteLengthUsedAsCharacterCount inspection.
package bytelengthusedascharactercount

import (
	"unicode/utf8"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ByteLengthUsedAsCharacterCount" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Count Unicode characters with the intended counting model."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "strlen")
	if call == nil {
		return
	}
	text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "string"))
	if known && utf8.ValidString(text) && utf8.RuneCountInString(text) < len(text) {
		ctx.ReportNode(call, message)
	}
}

func (rule) Semantic() {}
