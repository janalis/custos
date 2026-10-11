// Package authenticationcookieallowsscriptaccess implements the native AuthenticationCookieAllowsScriptAccess inspection.
package authenticationcookieallowsscriptaccess

import (
	"strings"

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

// setting returns the expression that decides the option, or nil when the
// call leaves it at its default. The second result is false when a literal
// options array cannot be read completely.
func setting(ctx *analysis.Context, c *syntax.FuncCall, pos int, name string) (syntax.Expr, bool) {
	a := semanticquery.NativeArray(ctx, semanticquery.CallArgument(c.Args, 2, "expires_or_options"))
	if a == nil {
		// PHP rejects an options array combined with the legacy arguments, so
		// an explicit legacy argument decides whatever the expiry expression is.
		return semanticquery.CallArgument(c.Args, pos, name), true
	}
	// PHP matches option names case-insensitively while walking the array,
	// so the last matching element in array order wins.
	var order []string
	values := map[string]syntax.Expr{}
	for _, item := range a.Items {
		if item.Key == nil {
			continue
		}
		key, ok := semanticquery.NativeArrayKey(ctx, item.Key)
		if !ok {
			return nil, false
		}
		if !strings.EqualFold(key, "s:"+name) {
			continue
		}
		if _, seen := values[key]; !seen {
			order = append(order, key)
		}
		values[key] = item.Value
	}
	if len(order) == 0 {
		return nil, true
	}
	return values[order[len(order)-1]], true
}
