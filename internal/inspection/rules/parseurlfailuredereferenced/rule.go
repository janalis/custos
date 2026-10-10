// Package parseurlfailuredereferenced implements the native ParseUrlFailureDereferenced inspection.
package parseurlfailuredereferenced

import (
	"net/url"
	"strconv"
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Reject URL parsing failure before indexing."

type rule struct{}

// New constructs the stateless inspection.
func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "ParseUrlFailureDereferenced" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(n) {
		return
	}
	a := n.(*syntax.ArrayDimFetch)
	if !astquery.ArrayFetchRequiresRead(a) {
		return
	}
	if !semanticquery.ExpansionDPristine(ctx, a.Var, n) {
		return
	}
	c, ok := semanticquery.NativeLocalValue(ctx, a.Var).(*syntax.FuncCall)
	if !ok || !semanticquery.NativeBuiltin(ctx, c, "parse_url") {
		return
	}
	component := semanticquery.CallArgument(c.Args, 1, "component")
	if component != nil {
		v, k := semanticquery.NativeInt(ctx, component)
		if !k || v != -1 {
			return
		}
	}
	if raw, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "url")); known && !literalURLFails(raw) {
		return
	}
	if !semanticquery.NativeLocalSentinelGuard(ctx, a.Var, "false") && !ctx.Flow().Excludes(a.Var, "false") {
		ctx.ReportNode(a, message)
	}
}

// Literal URL diagnostics require a supported failure proof. Unknown literal
// shapes cannot justify a missing runtime guard. Dynamic inputs may still fail.
func literalURLFails(raw string) bool {
	if lower := strings.ToLower(raw); lower == "http://" || lower == "https://" {
		return true
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return false
	}
	if port := u.Port(); port != "" {
		n, err := strconv.Atoi(port)
		return err != nil || n > 65535
	}
	return false
}
