// Package requestcontrolledsessionid implements the native RequestControlledSessionId inspection.
package requestcontrolledsessionid

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Generate session identifiers independently of request data."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "RequestControlledSessionId" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "session_id") {
		return
	}
	value := semanticquery.CallArgument(call.Args, 0, "id")
	if _, ok := request(ctx, value); ok {
		ctx.ReportNode(call, message)
	}
}

func request(ctx *analysis.Context, e syntax.Expr) (string, bool) {
	x, ok := semanticquery.NativeValue(ctx, e).(*syntax.ArrayDimFetch)
	if !ok {
		return "", false
	}
	v, ok := x.Var.(*syntax.Variable)
	if !ok {
		return "", false
	}
	switch v.Name {
	case "_GET", "_POST", "_REQUEST", "_COOKIE", "_SERVER":
	default:
		return "", false
	}
	if !semanticquery.NativeRequestUnwritten(ctx, x) {
		return "", false
	}
	key, known := semanticquery.NativeString(ctx, x.Dim)
	return key, known
}
