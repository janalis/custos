// Package corsoriginlistinvalid implements the native CorsOriginListInvalid inspection.
package corsoriginlistinvalid

import (
	"net/url"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Send one allowed origin in the CORS header."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CorsOriginListInvalid" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	key, value, _, known := semanticquery.ExpansionHeader(ctx, c)
	if !known || key != "access-control-allow-origin" {
		return
	}
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' })
	if len(parts) < 2 {
		return
	}
	for _, part := range parts {
		u, err := url.Parse(part)
		if err != nil || u.Scheme == "" || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || strings.ContainsAny(u.Host, "*,") {
			return
		}
	}
	ctx.ReportNode(c, message)
}
