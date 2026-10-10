// Package curlwritecallbackmissingbytecount implements the native CurlWriteCallbackMissingByteCount inspection.
package curlwritecallbackmissingbytecount

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Return the number of bytes handled by the write callback."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlWriteCallbackMissingByteCount" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "curl_setopt") {
		return
	}
	key, ok := semanticquery.CallArgument(c.Args, 1, "option").(*syntax.ConstFetch)
	if !ok || semanticquery.GlobalConstName(ctx, key) != "CURLOPT_WRITEFUNCTION" {
		return
	}
	callback, ok := semanticquery.CallArgument(c.Args, 2, "value").(*syntax.Closure)
	if !ok {
		return
	}
	values, complete := semanticquery.NativeReturns(callback.Body)
	if complete {
		for _, value := range values {
			constant, ok := syntax.UnwrapParens(value).(*syntax.ConstFetch)
			if !ok {
				return
			}
			name := semanticquery.GlobalConstName(ctx, constant)
			if name != "false" && name != "null" {
				return
			}
		}
		ctx.ReportNode(callback, message)
		return
	}
	for _, st := range callback.Body.Stmts {
		switch st := st.(type) {
		case *syntax.Echo:
		case *syntax.ExprStmt:
			if syntax.Terminates(st) {
				return
			}
		case *syntax.Return:
			if st.Expr != nil {
				return
			}
		default:
			return
		}
	}
	ctx.ReportNode(callback, message)
}
