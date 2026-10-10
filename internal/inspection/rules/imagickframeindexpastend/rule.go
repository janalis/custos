// Package imagickframeindexpastend implements the native ImagickFrameIndexPastEnd inspection.
package imagickframeindexpastend

import (
	"custos/internal/inspection/analysis"

	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Select an index below the image count."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ImagickFrameIndexPastEnd" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.ExpansionDUnaliased(ctx, c.Var, n) {
		return
	}
	if !semanticquery.NativeMethod(ctx, c, "Imagick", "setIteratorIndex") {
		return
	}
	e := semanticquery.CallArgument(c.Args, 0, "index")
	countCall, ok := semanticquery.NativeLocalValue(ctx, e).(*syntax.MethodCall)
	if ok && semanticquery.NativeMethod(ctx, countCall, "Imagick", "getNumberImages") && semanticquery.NativeSameValue(ctx, countCall.Var, c.Var) {
		for _, record := range ctx.Flow().Calls(syntax.EnclosingVariableScope(c)) {
			prior, ok := record.Node.(*syntax.MethodCall)
			if ok && prior.Span().Start > countCall.Span().End && prior.Span().End < c.Span().Start && semanticquery.NativeSameValue(ctx, prior.Var, c.Var) {
				return // The image list may have changed since the count was read.
			}
		}
		ctx.ReportNode(n, message)
		return
	}
	index, known := semanticquery.NativeContractInt(ctx, e)
	if !known || index < 0 {
		return
	}
	count := semanticquery.ExpansionDFrames(ctx, c)
	if count > 0 && index >= int64(count) {
		ctx.ReportNode(n, message)
	}
}
