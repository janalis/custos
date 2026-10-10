// Package utf8bytetruncation implements the Utf8ByteTruncation inspection.
package utf8bytetruncation

import (
	"unicode/utf8"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "Utf8ByteTruncation" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }

const message = "Keep UTF-8 code points intact when truncating text."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call, _ := semanticquery.GlobalCall(ctx, n, "substr")
	if call == nil {
		return
	}
	text, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "string"))
	start, ok := semanticquery.NativeInt(ctx, semanticquery.CallArgument(call.Args, 1, "offset"))
	if !known || !ok || !utf8.ValidString(text) {
		return
	}
	size := int64(len(text))
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
	if length := semanticquery.CallArgument(call.Args, 2, "length"); length != nil {
		count, ok := semanticquery.NativeInt(ctx, length)
		if !ok {
			return
		}
		if count < 0 {
			end = size + count
		} else if count < size-start {
			end = start + count
		}
	}
	if end < start {
		end = start
	}
	if !utf8.ValidString(text[start:end]) {
		ctx.ReportNode(call, message)
	}
}

func (rule) Semantic() {}
