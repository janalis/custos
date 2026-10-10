// Package truncationassumedtorewind implements the native TruncationAssumedToRewind inspection.
package truncationassumedtorewind

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Rewind after truncating the stream."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "TruncationAssumedToRewind" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "fwrite") && !semanticquery.NativeBuiltin(ctx, c, "fputs") {
		return
	}
	h := semanticquery.CallArgument(c.Args, 0, "stream")
	_, known := semanticquery.NativeStreamMode(ctx, h, c)
	if !known {
		return
	}
	position := int64(-1)
	truncated := false
	for _, n := range semanticquery.NativeStreamCalls(ctx, c, h, "fseek", "rewind", "fwrite", "fputs", "ftruncate") {
		p := n
		name := semanticquery.NativeBuiltinName(ctx, p)
		switch name {
		case "rewind":
			if semanticquery.NativeCallSucceeded(ctx, p, c) {
				position = 0
				truncated = false
			} else {
				return
			}
		case "fseek":
			offset, k := semanticquery.NativeInt(ctx, semanticquery.CallArgument(p.Args, 1, "offset"))
			w := semanticquery.CallArgument(p.Args, 2, "whence")
			whence := int64(0)
			wk := true
			if w != nil {
				whence, wk = semanticquery.NativeInt(ctx, w)
			}
			if !k || !wk || whence != 0 || !semanticquery.NativeCallSucceeded(ctx, p, c) {
				return
			}
			position = offset
			truncated = false
		case "fwrite", "fputs":
			value, k := semanticquery.NativeString(ctx, semanticquery.CallArgument(p.Args, 1, "data"))
			if !k || position < 0 || !semanticquery.NativeCallSucceeded(ctx, p, c) {
				return
			}
			position += int64(len(value))
		case "ftruncate":
			length, k := semanticquery.NativeInt(ctx, semanticquery.CallArgument(p.Args, 1, "size"))
			truncated = k && length == 0 && position > 0 && semanticquery.NativeCallSucceeded(ctx, p, c)
		}
	}
	if truncated {
		ctx.ReportNode(c, message)
	}
}
