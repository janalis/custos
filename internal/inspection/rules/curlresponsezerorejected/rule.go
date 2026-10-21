// Package curlresponsezerorejected implements CurlResponseZeroRejected.
package curlresponsezerorejected

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Compare cURL transport failure strictly with false."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlResponseZeroRejected" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c, ok := n.(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, c, "curl_exec") || semanticquery.FailureBranch(c) == nil {
		return
	}
	if curlResponseReturned(ctx, c, semanticquery.CallArgument(c.Args, 0, "handle")) {
		ctx.ReportNode(c, message)
	}
}

// curlResponseReturned replays every dominating option write on the handle in
// source order. Only a state whose last relevant write is a truthy
// RETURNTRANSFER proves that curl_exec() yields the body.
func curlResponseReturned(ctx *analysis.Context, at syntax.Node, handle syntax.Expr) bool {
	proven := false
	for _, event := range semanticquery.ScopeCallEvents(ctx, at) {
		c, ok := event.(*syntax.FuncCall)
		if !ok || c.Span().End >= at.Span().Start || !semanticquery.SameStableLocal(ctx, semanticquery.CallArgument(c.Args, 0, "handle"), handle) {
			continue
		}
		if !semanticquery.NativeDominates(c, at) {
			return false
		}
		switch semanticquery.NativeBuiltinName(ctx, c) {
		case "curl_setopt":
			proven = curlResponseWrite(ctx, proven, semanticquery.CallArgument(c.Args, 1, "option"), semanticquery.CallArgument(c.Args, 2, "value"))
		case "curl_setopt_array":
			options := semanticquery.NativeArray(ctx, semanticquery.CallArgument(c.Args, 1, "options"))
			if options == nil {
				proven = false
				continue
			}
			for _, item := range options.Items {
				proven = curlResponseWrite(ctx, proven, item.Key, item.Value)
			}
		default:
			return false
		}
	}
	return proven
}

// curlResponseWrite applies one option write. An option that is not a builtin
// constant may be anything, and a replaced write sink makes curl_exec() return
// a boolean again.
func curlResponseWrite(ctx *analysis.Context, proven bool, option, value syntax.Expr) bool {
	name := ""
	if c, ok := syntax.UnwrapParens(option).(*syntax.ConstFetch); ok {
		name = semanticquery.GlobalConstName(ctx, c)
		if symbol := ctx.Index().Constant(name, ctx.PHP); symbol == nil || !symbol.Builtin {
			name = ""
		}
	}
	switch name {
	case "CURLOPT_RETURNTRANSFER":
		enabled, known := semanticquery.NativeTruth(ctx, value)
		return known && enabled
	case "", "CURLOPT_FILE", "CURLOPT_WRITEFUNCTION":
		return false
	}
	return proven
}
