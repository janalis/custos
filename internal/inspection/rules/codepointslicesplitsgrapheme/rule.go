// Package codepointslicesplitsgrapheme implements the native CodePointSliceSplitsGrapheme inspection.
package codepointslicesplitsgrapheme

import (
	"strings"
	"unicode/utf8"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

const message = "Slice display text at grapheme boundaries."

// clusterBreakers lists the code points after which a combining mark starts
// its own grapheme cluster (UAX #29 CR, LF and Control, Unicode 17). Format
// characters that extend or prepend a cluster (ZWNJ, ZWJ, Arabic number signs,
// emoji tag characters) are deliberately absent.
var clusterBreakers = [...][2]rune{
	{0x0000, 0x001f},
	{0x007f, 0x009f},
	{0x00ad, 0x00ad},
	{0x061c, 0x061c},
	{0x180e, 0x180e},
	{0x200b, 0x200b},
	{0x200e, 0x200f},
	{0x2028, 0x202e},
	{0x2060, 0x206f},
	{0xfeff, 0xfeff},
	{0xfff0, 0xfffb},
	{0x13430, 0x1343f},
	{0x1bca0, 0x1bca3},
	{0x1d173, 0x1d17a},
	{0xe0000, 0xe001f},
	{0xe0080, 0xe00ff},
	{0xe01f0, 0xe0fff},
}

// breaksCluster reports whether r ends its grapheme cluster unconditionally.
func breaksCluster(r rune) bool {
	for _, span := range clusterBreakers {
		if r <= span[1] {
			return r >= span[0]
		}
	}
	return false
}

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
		return i > 0 && i < size && runes[i] >= 0x300 && runes[i] <= 0x36f && !breaksCluster(runes[i-1])
	}
	if boundary(start) || boundary(end) {
		ctx.ReportNode(c, message)
	}
}
