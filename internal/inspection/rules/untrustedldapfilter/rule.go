// Package untrustedldapfilter implements the native UntrustedLdapFilter inspection.
package untrustedldapfilter

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Escape request data for LDAP filter syntax."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "UntrustedLdapFilter" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	switch semanticquery.NativeBuiltinName(ctx, call) {
	case "ldap_search", "ldap_list", "ldap_read":
	default:
		return
	}
	filter := semanticquery.CallArgument(call.Args, 2, "filter")
	if untrusted(ctx, filter, 0) {
		ctx.ReportNode(filter, message)
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

func untrusted(ctx *analysis.Context, e syntax.Expr, depth int) bool {
	if depth > 16 {
		return false
	}
	if _, ok := request(ctx, e); ok {
		return true
	}
	b, ok := semanticquery.NativeValue(ctx, e).(*syntax.Binary)
	if ok && b.Op.Kind == syntax.TDot {
		return untrusted(ctx, b.Left, depth+1) || untrusted(ctx, b.Right, depth+1)
	}
	return false
}
