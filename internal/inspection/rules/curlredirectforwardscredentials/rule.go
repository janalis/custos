// Package curlredirectforwardscredentials implements the native CurlRedirectForwardsCredentials inspection.
package curlredirectforwardscredentials

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Restrict credential forwarding across redirects."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlRedirectForwardsCredentials" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "curl_exec") {
		return
	}
	h := semanticquery.CallArgument(c.Args, 0, "handle")
	follow, fk := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_FOLLOWLOCATION")
	unrestricted, uk := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_UNRESTRICTED_AUTH")
	credentials, ck := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_USERPWD")
	fv, fn := semanticquery.NativeTruth(ctx, follow)
	uv, un := semanticquery.NativeTruth(ctx, unrestricted)
	credentialText, cn := semanticquery.NativeString(ctx, credentials)
	if fk && uk && ck && fn && un && fv && uv && cn && credentialText != "" {
		arg := unrestricted.Parent().(*syntax.Arg)
		call := arg.Parent().Parent().(*syntax.FuncCall)
		ctx.ReportNode(semanticquery.CallArgument(call.Args, 1, "option"), message)
	}
}
