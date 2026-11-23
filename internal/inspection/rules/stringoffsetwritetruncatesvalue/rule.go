// Package stringoffsetwritetruncatesvalue implements the native StringOffsetWriteTruncatesValue inspection.
package stringoffsetwritetruncatesvalue

import (
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Assign one byte to a string offset."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "StringOffsetWriteTruncatesValue" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KAssign} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.File.Recovered() || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	a := n.(*syntax.Assign)
	if a.Op.Kind != syntax.TEqual || a.ByRef {
		return
	}
	o, ok := syntax.UnwrapParens(a.Var).(*syntax.ArrayDimFetch)
	if !ok {
		return
	}
	_, known := semanticquery.NativeInt(ctx, o.Dim)
	value, k := semanticquery.NativeString(ctx, a.Value)
	if known && k && len(value) > 1 && ctx.TypeOf(o.Var).OnlyOf("string") {
		ctx.ReportNode(a, message)
	}
}
