// Package curltransportsuccessashttpsuccess implements the native CurlTransportSuccessAsHttpSuccess inspection.
package curltransportsuccessashttpsuccess

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Check the HTTP response status before reporting success."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlTransportSuccessAsHttpSuccess" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	call := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, call, "curl_exec") {
		return
	}
	handle := semanticquery.CallArgument(call.Args, 0, "handle")
	origin, known := ctx.Flow().Resolve(handle)
	if !known {
		return
	}
	init, ok := origin.(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, init, "curl_init") {
		return
	}
	url, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(init.Args, 0, "url"))
	if !known || (!strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://")) {
		return
	}
	if !semanticquery.NativeCurlDefault(ctx, call, handle, "CURLOPT_FAILONERROR") {
		return
	}
	p, _ := astquery.ParentSkipParens(call)
	b, ok := p.(*syntax.Binary)
	if !ok || b.Op.Kind != syntax.TIsNotIdentical {
		return
	}
	truth, known := astquery.BoolConst(b.Right)
	if !known || truth {
		return
	}
	parent, _ := astquery.ParentSkipParens(b)
	condition, ok := parent.(*syntax.If)
	if !ok || condition.Else != nil {
		return
	}
	body, ok := condition.Body.(*syntax.Block)
	if !ok || len(body.Stmts) != 1 {
		return
	}
	ret, ok := body.Stmts[0].(*syntax.Return)
	if !ok {
		return
	}
	truth, known = astquery.BoolConst(ret.Expr)
	if known && truth {
		ctx.ReportNode(call, message)
	}
}
