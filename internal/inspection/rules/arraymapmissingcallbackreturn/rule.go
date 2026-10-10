// Package arraymapmissingcallbackreturn implements the native ArrayMapMissingCallbackReturn inspection.
package arraymapmissingcallbackreturn

import (
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return the mapped value from this callback."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ArrayMapMissingCallbackReturn" }
func (rule) Semantic()                {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "array_map") {
		return
	}
	cl, ok := semanticquery.NativeValue(ctx, semanticquery.CallArgument(call.Args, 0, "callback")).(*syntax.Closure)
	if !ok || cl.Body == nil {
		return
	}
	if cl.ReturnType != nil && strings.EqualFold(ctx.Text(cl.ReturnType), "void") {
		return
	}
	returned, transformed := false, false
	syntax.Inspect(cl.Body, func(x syntax.Node) bool {
		if !semanticquery.NativeCallbackReachable(x) {
			return false
		}
		switch z := x.(type) {
		case *syntax.Closure, *syntax.ArrowFunction, *syntax.Function, *syntax.ClassLike:
			return false
		case *syntax.Return:
			if z.Expr != nil {
				returned = true
			}
		case *syntax.Yield, *syntax.YieldFrom:
			returned = true
		case *syntax.ExprStmt:
			if fc, ok := z.Expr.(*syntax.FuncCall); ok {
				switch semanticquery.NativeBuiltinName(ctx, fc) {
				case "trim", "ltrim", "rtrim", "strtolower", "strtoupper", "substr", "str_replace", "mb_strtolower", "mb_strtoupper":
					transformed = true
				}
			}
		}
		return true
	})
	if returned || !transformed {
		return
	}
	var fixes []diagnostic.Fix
	if len(cl.Body.Stmts) == 1 && cl.ReturnType == nil {
		if st, ok := cl.Body.Stmts[0].(*syntax.ExprStmt); ok {
			fixes = append(fixes, diagnostic.Fix{Title: "Return the mapped value", Edits: func() []diagnostic.TextEdit {
				return []diagnostic.TextEdit{{Span: syntax.Span{Start: st.Span().Start, End: st.Span().Start}, NewText: "return "}}
			}})
		}
	}
	ctx.ReportNode(call, message, fixes...)
}
