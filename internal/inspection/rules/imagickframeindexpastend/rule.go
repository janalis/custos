// Package imagickframeindexpastend implements the native ImagickFrameIndexPastEnd inspection.
package imagickframeindexpastend

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
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
		if !countUnchanged(ctx, e) {
			return
		}
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

// countUnchanged follows e through plain local copies and requires the latest
// earlier binding of every variable on the way to be a simple assignment:
// `$n--`, a foreach or destructuring target, a reference binding and similar
// rebinds mean the variable no longer holds the count that was read.
func countUnchanged(ctx *analysis.Context, e syntax.Expr) bool {
	scope := syntax.EnclosingVariableScope(e)
	for {
		v, ok := syntax.UnwrapParens(e).(*syntax.Variable)
		if !ok {
			return true
		}
		var last flowquery.VarAccess
		for _, access := range flowquery.VarAccesses(ctx.File, scope, v.Name) {
			if access.Var.Span().Start >= v.Span().Start {
				break
			}
			if a, ok := access.By.(*syntax.Assign); ok && a.Value.Span().End >= v.Span().Start {
				continue // The assignment being evaluated does not bind its own operand.
			}
			if access.Write {
				last = access
			}
		}
		a, ok := last.By.(*syntax.Assign)
		if !ok || a.Var != syntax.Expr(last.Var) || a.ByRef || a.Op.Kind != syntax.TEqual {
			return false
		}
		e = a.Value
	}
}
