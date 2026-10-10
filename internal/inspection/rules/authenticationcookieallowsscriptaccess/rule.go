// Package authenticationcookieallowsscriptaccess implements the native AuthenticationCookieAllowsScriptAccess inspection.
package authenticationcookieallowsscriptaccess

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Enable HttpOnly for the authentication cookie."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "AuthenticationCookieAllowsScriptAccess" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	name := semanticquery.NativeBuiltinName(ctx, call)
	if name != "setcookie" && name != "setrawcookie" {
		return
	}
	cookie, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 0, "name"))
	if !known {
		return
	}
	auth := false
	for _, name := range ctx.List("authenticationCookies") {
		if name == cookie {
			auth = true
		}
	}
	if !auth {
		return
	}
	http, ok := setting(ctx, call, 6, "httponly")
	if !ok || http == nil {
		return
	}
	value, known := semanticquery.NativeTruth(ctx, http)
	if known && !value {
		ctx.ReportNode(http, message)
	}
}

func setting(ctx *analysis.Context, c *syntax.FuncCall, pos int, name string) (syntax.Expr, bool) {
	value := semanticquery.CallArgument(c.Args, 2, "expires_or_options")
	if value == nil {
		return semanticquery.CallArgument(c.Args, pos, name), true
	}
	if a := semanticquery.NativeArray(ctx, value); a != nil {
		entries, ok := semanticquery.NativeArrayEntries(ctx, value)
		if !ok {
			return nil, false
		}
		return entries["s:"+name], true
	}
	if _, ok := semanticquery.NativeInt(ctx, value); ok {
		return semanticquery.CallArgument(c.Args, pos, name), true
	}
	return nil, false
}
