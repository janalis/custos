// Package codepointslicesplitsgrapheme implements the native CodePointSliceSplitsGrapheme inspection.
package codepointslicesplitsgrapheme

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Slice display text at grapheme boundaries."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CodePointSliceSplitsGrapheme" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if ctx.PHP < phpversion.PHP70 || !semanticquery.NativeBuiltin(ctx, c, "mb_substr") {
		return
	}
	text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "string"))
	if !known || !utf8.ValidString(text) {
		return
	}
	encoding, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 3, "encoding"))
	if !known || !strings.EqualFold(encoding, "UTF-8") {
		return
	}
	runes := []rune(text)
	size := int64(len(runes))
	start, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(c.Args, 1, "start"))
	if !known {
		return
	}
	if start < 0 {
		start += size
	}
	if start < 0 {
		start = 0
	}
	if start > size {
		return
	}
	end := size
	if length := semanticquery.CallArgument(c.Args, 2, "length"); length != nil {
		count, known := semanticquery.NativeInt(ctx, length)
		if !known {
			return
		}
		if count < 0 {
			end = size + count
		} else if count < size-start {
			end = start + count
		}
	}
	if end <= start {
		return
	}
	boundary := func(i int64) bool {
		return i > 0 && i < size && runes[i] >= 0x300 && runes[i] <= 0x36f && !unicode.IsControl(runes[i-1])
	}
	if boundary(start) || boundary(end) {
		ctx.ReportNode(c, message)
	}
}
