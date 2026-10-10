// Package untrustedxmlwriterrawcontent implements UntrustedXmlWriterRawContent.
package untrustedxmlwriterrawcontent

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Write untrusted XML content as escaped text."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UntrustedXmlWriterRawContent" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "XMLWriter", "writeRaw") {
		return
	}
	source, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(c.Args, 0, "content")).(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	root, ok := source.Var.(*syntax.Variable)
	if !ok || !semanticquery.NativeRequestUnwritten(ctx, source) {
		return
	}
	if root.Name == "_GET" || root.Name == "_POST" || root.Name == "_REQUEST" || root.Name == "_COOKIE" {
		ctx.ReportNode(c, message)
	}
}
