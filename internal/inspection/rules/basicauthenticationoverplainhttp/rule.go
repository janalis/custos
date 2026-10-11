// Package basicauthenticationoverplainhttp implements the native BasicAuthenticationOverPlainHttp inspection.
package basicauthenticationoverplainhttp

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Protect basic authentication with HTTPS."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "BasicAuthenticationOverPlainHttp" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "curl_setopt") {
		return
	}
	option, ok := syntax.UnwrapParens(semanticquery.CallArgument(call.Args, 1, "option")).(*syntax.ConstFetch)
	if !ok || semanticquery.GlobalConstName(ctx, option) != "CURLOPT_USERPWD" {
		return
	}
	handle := semanticquery.CallArgument(call.Args, 0, "handle")
	creation, ok := semanticquery.NativeValue(ctx, handle).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, creation, "curl_init") {
		return
	}
	stmt, ok := call.Parent().(*syntax.ExprStmt)
	if !ok {
		return
	}
	next, ok := astquery.NextStmt(ctx.File, stmt)
	if !ok {
		return
	}
	st, ok := next.(*syntax.ExprStmt)
	if !ok {
		return
	}
	exec, ok := st.Expr.(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, exec, "curl_exec") {
		return
	}
	other := semanticquery.CallArgument(exec.Args, 0, "handle")
	a, b := ctx.Flow().Value(handle), ctx.Flow().Value(other)
	if !a.Complete || !b.Complete || a.Identity == 0 || a.Identity != b.Identity || a.Invalidated != b.Invalidated {
		return
	}
	credentials, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(call.Args, 2, "value"))
	if !known || credentials == "" {
		return
	}
	// An explicit CURLOPT_URL replaces whatever curl_init received; the
	// curl_init argument only matters while the option keeps its default.
	target := semanticquery.CallArgument(creation.Args, 0, "url")
	if configured, explicit := semanticquery.NativeCurlOption(ctx, exec, handle, "CURLOPT_URL"); explicit {
		target = configured
	} else if !semanticquery.NativeCurlDefault(ctx, exec, handle, "CURLOPT_URL") {
		return
	}
	url, known := semanticquery.NativeString(ctx, target)
	if !known {
		return
	}
	if !strings.HasPrefix(strings.ToLower(url), "http://") {
		return
	}
	if auth, explicit := semanticquery.NativeCurlOption(ctx, exec, handle, "CURLOPT_HTTPAUTH"); explicit {
		value, known := semanticquery.NativeInt(ctx, auth)
		if !known || value != 1 {
			return
		}
	} else if !semanticquery.NativeCurlDefault(ctx, exec, handle, "CURLOPT_HTTPAUTH") {
		return
	}
	ctx.ReportNode(call, message)
}
