// Package zipentrynamecollision implements the native ZipEntryNameCollision inspection.
package zipentrynamecollision

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Make ZIP entry replacement explicit."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipEntryNameCollision" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "ZipArchive", "addFromString") {
		return
	}
	name, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "name"))
	if !known || semanticquery.CallArgument(c.Args, 2, "flags") != nil {
		return
	}
	for p := c.Parent(); p != nil; p = p.Parent() {
		branch, ok := p.(*syntax.If)
		if !ok || !branch.Body.Span().Contains(c.Span()) {
			continue
		}
		add, ok := syntax.UnwrapParens(branch.Cond).(*syntax.MethodCall)
		if !ok || add == c || !semanticquery.NativeMethod(ctx, add, "ZipArchive", "addFromString") || !semanticquery.ExpansionCSame(ctx, add.Var, c.Var) {
			continue
		}
		previous, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(add.Args, 0, "name"))
		if known && previous == name && len(semanticquery.NativePriorCalls(ctx, c, c.Var, "close", "deleteName")) == 0 {
			ctx.ReportNode(c, message)
			return
		}
	}
}
