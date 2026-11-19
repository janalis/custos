// Package corssubdomainwildcardinvalid implements the native CorsSubdomainWildcardInvalid inspection.
package corssubdomainwildcardinvalid

import (
	"strings"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/flowquery"
	"github.com/janalis/custos/internal/inspection/semanticquery"
	"github.com/janalis/custos/internal/php/syntax"
)

const message = "Send a concrete CORS origin."

type rule struct{}

// New returns the CorsSubdomainWildcardInvalid inspection.
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
	key, value, _, known := semanticquery.LiteralHeader(ctx, c)
	if known && key == "access-control-allow-origin" && value != "*" && strings.Contains(value, "*") {
		ctx.ReportNode(c, message)
	}
}
