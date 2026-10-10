// Package zipsourceremovedbeforearchiveclose implements the native ZipSourceRemovedBeforeArchiveClose inspection.
package zipsourceremovedbeforearchiveclose

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Keep ZIP source files until the archive closes."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipSourceRemovedBeforeArchiveClose" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "unlink") {
		return
	}
	path, ok := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "filename"))
	if !ok {
		return
	}
	for p := c.Parent(); p != nil; p = p.Parent() {
		branch, ok := p.(*syntax.If)
		if !ok || !branch.Body.Span().Contains(c.Span()) {
			continue
		}
		add, ok := syntax.UnwrapParens(branch.Cond).(*syntax.MethodCall)
		if !ok || !semanticquery.NativeMethod(ctx, add, "ZipArchive", "addFile") {
			continue
		}
		source, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(add.Args, 0, "filepath"))
		if !known || source != path {
			continue
		}
		flags := semanticquery.CallArgument(add.Args, 4, "flags")
		if flags != nil {
			continue
		}
		if len(semanticquery.NativePriorCalls(ctx, c, add.Var, "close")) > 0 {
			return
		}
		ctx.ReportNode(c, message)
		return
	}
}
