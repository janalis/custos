// Package mysqliescapedvaluesurvivescharsetchange implements MysqliEscapedValueSurvivesCharsetChange.
package mysqliescapedvaluesurvivescharsetchange

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Escape values after selecting the connection charset."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "MysqliEscapedValueSurvivesCharsetChange" }
func (rule) Semantic()                {}
func (rule) Flow()                    {}
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethodCall} }
func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	check(ctx, n, message)
}

func check(ctx *analysis.Context, n syntax.Node, message string) {
	if len(ctx.File.Errors) > 0 || !semanticquery.NativeCallbackReachable(ctx, n) {
		return
	}
	c := n.(*syntax.MethodCall)
	if !semanticquery.NativeMethod(ctx, c, "mysqli", "query") {
		return
	}
	var escaped []*syntax.MethodCall
	sql := semanticquery.CallArgument(c.Args, 0, "query")
	if sql == nil {
		return
	}
	syntax.Inspect(sql, func(n syntax.Node) bool {
		e, ok := n.(syntax.Expr)
		if !ok {
			return true
		}
		origin, ok := semanticquery.NativeLocalValue(ctx, e).(*syntax.MethodCall)
		if ok && semanticquery.NativeMethod(ctx, origin, "mysqli", "real_escape_string") {
			escaped = append(escaped, origin)
		}
		return true
	})
	for _, escape := range escaped {
		if !semanticquery.SameStableLocal(ctx, c.Var, escape.Var) {
			continue
		}
		// Only the last selection on each side decides: an earlier change that
		// a later call overrides, restores or leaves unproven is not evidence.
		prior := expansionDLastCharsetCall(ctx, escape, escape.Var, 0)
		if prior == nil || (!semanticquery.NativeDominates(prior, escape) && !semanticquery.NativeCallSucceeded(ctx, prior, escape)) {
			continue
		}
		before, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(prior.Args, 0, "charset"))
		if !known {
			continue
		}
		change := expansionDLastCharsetCall(ctx, c, c.Var, escape.Span().End)
		if change == nil || !semanticquery.NativeCallSucceeded(ctx, change, c) {
			continue
		}
		after, known := semanticquery.NativeString(ctx, semanticquery.CallArgument(change.Args, 0, "charset"))
		if known && !strings.EqualFold(before, after) {
			ctx.ReportNode(c, message)
			return
		}
	}
}

// expansionDLastCharsetCall returns the last mysqli::set_charset call on the
// receiver written after offset from and before at, whether or not it is
// proven to execute; callers decide what that call proves.
func expansionDLastCharsetCall(ctx *analysis.Context, at syntax.Node, receiver syntax.Expr, from uint32) *syntax.MethodCall {
	var last *syntax.MethodCall
	for _, event := range semanticquery.ScopeCallEvents(ctx, at) {
		c, ok := event.(*syntax.MethodCall)
		if ok && c.Span().Start > from && c.Span().End < at.Span().Start && semanticquery.NativeMethod(ctx, c, "mysqli", "set_charset") && semanticquery.SameStableLocal(ctx, c.Var, receiver) {
			last = c
		}
	}
	return last
}
