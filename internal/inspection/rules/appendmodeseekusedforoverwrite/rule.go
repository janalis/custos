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
	if seek && !emptied(ctx, c, h) {
		ctx.ReportNode(c, message)
	}
}

// emptied reports whether an earlier ftruncate may have left the stream empty,
// which makes offset 0 the end of file. Conditional truncations count too. A
// later dominating write of a non-empty literal restores the proof.
func emptied(ctx *analysis.Context, at *syntax.FuncCall, h syntax.Expr) bool {
	cut, found := uint32(0), false
	scope := syntax.EnclosingVariableScope(at)
	visit := func(n syntax.Node) bool {
		if n != scope && syntax.IsVariableScope(n) {
			return false
		}
		t, ok := n.(*syntax.FuncCall)
		if !ok || t.Span().End > at.Span().Start || !semanticquery.NativeBuiltin(ctx, t, "ftruncate") || !semanticquery.NativeSameValue(ctx, semanticquery.CallArgument(t.Args, 0, "stream"), h) {
			return true
		}
		if size, known := semanticquery.NativeInt(ctx, semanticquery.CallArgument(t.Args, 1, "size")); !known || size == 0 {
			cut, found = t.Span().Start, true
		}
		return true
	}
	if scope == nil {
		syntax.InspectFile(ctx.File, visit)
	} else {
		syntax.Inspect(scope, visit)
	}
	if !found {
		return false
	}
	for _, w := range semanticquery.NativeStreamCalls(ctx, at, h, "fwrite", "fputs") {
		if data, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(w.Args, 1, "data")); known && data != "" && w.Span().Start > cut {
			return false
		}
	}
	return true
}
