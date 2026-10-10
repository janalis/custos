// Package corssubdomainwildcardinvalid implements the native CorsSubdomainWildcardInvalid inspection.
package corssubdomainwildcardinvalid

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/flowquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Send a concrete CORS origin."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "CorsSubdomainWildcardInvalid" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if !flowquery.Reachable(n, syntax.EnclosingFuncLike(n)) {
		return
	}
	c := n.(*syntax.FuncCall)
	key, value, _, known := semanticquery.ExpansionHeader(ctx, c)
	if known && key == "access-control-allow-origin" && value != "*" && strings.Contains(value, "*") {
		ctx.ReportNode(c, message)
	}
}
