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
	"custos/internal/semantic/types"
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
	// Only a file spelling it can reach parse_url (see astquery.Mentions).
	if !astquery.Mentions(ctx.File, "parse_url") {
		return
	}
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	a := n.(*syntax.ArrayDimFetch)
	if !astquery.ArrayFetchRequiresRead(a) {
		return
	}
	if !semanticquery.ExpansionDPristine(ctx, a.Var, n) && !semanticquery.NativeFailureFlowOnly(ctx, a.Var) {
		return
	}
	// Among several reaching assignments only a parse_url result counts:
	// an unguarded result of another call does not stand in for it.
	var c *syntax.FuncCall
	for _, candidate := range semanticquery.NativeLocalBuiltinResults(ctx, a.Var, "false", "parse_url") {
		if call, ok := candidate.(*syntax.FuncCall); ok && relevantParse(ctx, call) {
			c = call
			break
		}
	}
	ok := c != nil
	if !ok {
		// The lexical scan only reads assignments that are statements; the
		// flow also knows a result assigned inside a condition.
		c, ok = semanticquery.NativeValue(ctx, a.Var).(*syntax.FuncCall)
	}
	if !ok || !semanticquery.NativeBuiltin(ctx, c, "parse_url") {
		return
	}
	if !relevantParse(ctx, c) {
		return
	}

	if semanticquery.NativeFailureExcludedSingle(ctx, a.Var, types.Of("array", "false"), "false") {
		return
	}
	ctx.ReportNode(a, message)
}

// relevantParse excludes other result contracts before choosing a producer,
// so a safe literal or component result cannot hide another failing parse.
func relevantParse(ctx *analysis.Context, c *syntax.FuncCall) bool {
	if component, state := semanticquery.CallArgumentBinding(c.Args, 1, "component"); state != semanticquery.ArgumentMissing {
		v, known := semanticquery.NativeInt(ctx, component)
		if !known || v != -1 {
			return false
		}
	}
	raw, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(c.Args, 0, "url"))
	return !known || literalURLFails(raw)
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
