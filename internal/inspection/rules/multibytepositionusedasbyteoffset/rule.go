// Package multibytepositionusedasbyteoffset implements the native MultibytePositionUsedAsByteOffset inspection.
package multibytepositionusedasbyteoffset

import (
	"strings"
	"unicode/utf8"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use multibyte offsets with multibyte slicing."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MultibytePositionUsedAsByteOffset" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "substr") {
		return
	}
	source := semanticquery.CallArgument(c.Args, 0, "string")
	text, known := semanticquery.NativeString(ctx, source)
	if !known || !utf8.ValidString(text) {
		return
	}
	position, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 1, "offset")).(*syntax.FuncCall)
	if !ok {
		return
	}
	name := semanticquery.NativeBuiltinName(ctx, position)
	if name != "mb_strpos" && name != "mb_stripos" {
		return
	}
	same, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(position.Args, 0, "haystack"))
	if !known || same != text {
		return
	}
	encoding, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(position.Args, 3, "encoding"))
	if !known || !strings.EqualFold(encoding, "UTF-8") {
		return
	}
	needle, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(position.Args, 1, "needle"))
	if !known || needle == "" {
		return
	}
	if offset := semanticquery.CallArgument(position.Args, 2, "offset"); offset != nil {
		value, known := semanticquery.NativeInt(ctx, offset)
		if !known || value != 0 {
			return
		}
	}
	if name == "mb_stripos" { // Only ASCII case folding is proven compatible here.
		for _, ch := range needle {
			if ch > 127 {
				return
			}
		}
		needle = strings.ToLower(needle)
		text = strings.ToLower(text)
	}
	index := strings.Index(text, needle)
	if index < 0 || utf8.RuneCountInString(text[:index]) == index {
		return
	}
	ctx.ReportNode(c, message)
}
