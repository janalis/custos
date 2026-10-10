// Package curlheaderlistoverwritten implements the native CurlHeaderListOverwritten inspection.
package curlheaderlistoverwritten

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Combine HTTP headers before setting the final list."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CurlHeaderListOverwritten" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if !semanticquery.NativeBuiltin(ctx, c, "curl_exec") {
		return
	}
	h := semanticquery.CallArgument(c.Args, 0, "handle")
	latest, known := semanticquery.NativeCurlOption(ctx, c, h, "CURLOPT_HTTPHEADER")
	if !known {
		return
	}
	auth, known := authorization(ctx, latest)
	if !known || auth {
		return
	}
	var earlier syntax.Expr
	for _, node := range semanticquery.NativePriorCalls(ctx, c, h, "curl_setopt") {
		setter := node.(*syntax.FuncCall)
		key, ok := semanticquery.CallArgument(setter.Args, 1, "option").(*syntax.ConstFetch)
		if ok && semanticquery.GlobalConstName(ctx, key) == "CURLOPT_HTTPHEADER" {
			value := semanticquery.CallArgument(setter.Args, 2, "value")
			if value == latest {
				break
			}
			earlier = value
		}
	}
	auth, known = authorization(ctx, earlier)
	if known && auth {
		ctx.ReportNode(latest, message)
	}
}

func authorization(ctx *analysis.Context, e syntax.Expr) (bool, bool) {
	entries, known := semanticquery.NativeArrayEntries(ctx, e)
	if !known || len(entries) == 0 {
		return false, false
	}
	found := false
	for _, value := range entries {
		s, known := semanticquery.NativeString(ctx, value)
		if !known {
			return false, false
		}
		parts := strings.SplitN(s, ":", 2)
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), "Authorization") {
			found = true
		}
	}
	return found, true
}
