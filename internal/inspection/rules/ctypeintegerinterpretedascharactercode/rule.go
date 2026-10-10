// Package ctypeintegerinterpretedascharactercode implements the native CtypeIntegerInterpretedAsCharacterCode inspection.
package ctypeintegerinterpretedascharactercode

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Pass a string to avoid character-code interpretation."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CtypeIntegerInterpretedAsCharacterCode" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "ctype_digit") {
		return
	}
	input := semanticquery.CallArgument(c.Args, 0, "text")
	integer, known := semanticquery.NativeInt(ctx, input)
	if known && integer >= 0 && integer <= 255 {
		ctx.ReportNode(input, message)
	}
}
