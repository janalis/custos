// Package xmlwritermemoryoutputinurimode implements the native XmlWriterMemoryOutputInUriMode inspection.
package xmlwritermemoryoutputinurimode

import (
	"github.com/janalis/custos/internal/inspection/analysis"

	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Use URI output for a URI-backed XML writer."

type rule struct{}

// New returns the XmlWriterMemoryOutputInUriMode inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "XmlWriterMemoryOutputInUriMode" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.File.Recovered() || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.UnaliasedLocal(ctx, c.Var, n) {
		return
	}
	if !semanticquery.NativeMethod(ctx, c, "XMLWriter", "outputMemory") {
		return
	}
	if semanticquery.NativeConstruction(ctx, c.Var, "XMLWriter") == nil {
		return
	}
	uri := false
	for _, p := range semanticquery.ObjectMethodHistory(ctx, c, c.Var) {
		if semanticquery.NativeMethod(ctx, p, "XMLWriter", "openUri") {
			uri = semanticquery.NativeCallSucceeded(ctx, p, c)
		}
		if semanticquery.NativeMethod(ctx, p, "XMLWriter", "openMemory") {
			uri = false
		}
	}
	if uri {
		ctx.ReportNode(n, message)
	}
}
