// Package appendmodeseekusedforoverwrite implements the native AppendModeSeekUsedForOverwrite inspection.
package appendmodeseekusedforoverwrite

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use an overwrite-capable stream mode."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AppendModeSeekUsedForOverwrite" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "fwrite") && !semanticquery.NativeBuiltin(ctx, c, "fputs") {
		return
	}
	h := semanticquery.CallArgument(c.Args, 0, "stream")
	mode, known := semanticquery.NativeStreamMode(ctx, h, c)
	if !known || mode == "" || mode[0] != 'a' {
		return
	}
	seek := false
	for _, n := range semanticquery.NativeStreamCalls(ctx, c, h, "fseek", "rewind") {
		p := n
		if semanticquery.NativeBuiltin(ctx, p, "rewind") {
			seek = semanticquery.NativeCallSucceeded(ctx, p, c)
			continue
		}
		offset, k := semanticquery.NativeInt(ctx, semanticquery.CallArgument(p.Args, 1, "offset"))
		whence := semanticquery.CallArgument(p.Args, 2, "whence")
		w := int64(0)
		wk := true
		if whence != nil {
			w, wk = semanticquery.NativeInt(ctx, whence)
		}
		seek = k && wk && offset == 0 && w == 0 && semanticquery.NativeCallSucceeded(ctx, p, c)
	}
	if seek {
		ctx.ReportNode(c, message)
	}
}
