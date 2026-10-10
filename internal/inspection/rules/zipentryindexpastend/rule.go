// Package zipentryindexpastend implements the native ZipEntryIndexPastEnd inspection.
package zipentryindexpastend

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Use an existing ZIP entry index."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ZipEntryIndexPastEnd" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.MethodCall)
	valid := false
	for _, method := range []string{"getNameIndex", "statIndex", "getFromIndex", "deleteIndex", "getStreamIndex"} {
		valid = valid || semanticquery.NativeMethod(ctx, c, "ZipArchive", method)
	}
	if !valid {
		return
	}
	arg := semanticquery.CallArgument(c.Args, 0, "index")
	if v, ok := semanticquery.NativeInt(ctx, arg); ok && v < 0 {
		ctx.ReportNode(c, message)
		return
	}
	property, ok := syntax.UnwrapParens(arg).(*syntax.PropertyFetch)
	if !ok || !semanticquery.ExpansionCSame(ctx, property.Var, c.Var) {
		return
	}
	name, ok := property.Name.(*syntax.Identifier)
	if ok && name.Value == "numFiles" {
		ctx.ReportNode(c, message)
	}
}
