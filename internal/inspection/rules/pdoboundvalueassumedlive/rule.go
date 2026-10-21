// Package pdoboundvalueassumedlive implements PdoBoundValueAssumedLive.
package pdoboundvalueassumedlive

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
)

const message = "Rebind the changed value before execution."

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "PdoBoundValueAssumedLive" }
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
	_, paramsState := semanticquery.CallArgumentBinding(c.Args, 0, "params")
	if !semanticquery.NativeMethod(ctx, c, "PDOStatement", "execute") || paramsState != semanticquery.ArgumentMissing {
		return
	}
	bindings := semanticquery.EarlierMethodCalls(ctx, c, c.Var, "PDOStatement", "bindvalue", "bindparam")
	for i, binding := range bindings {
		if !semanticquery.NativeMethod(ctx, binding, "PDOStatement", "bindValue") {
			continue
		}
		variable, ok := semanticquery.CallArgument(binding.Args, 1, "value").(*syntax.Variable)
		if !ok {
			return
		}
		typ := ctx.TypeOf(variable)
		if !typ.OnlyOf("int", "string", "float", "bool", "null") {
			continue
		}
		rebound := false
		for _, later := range bindings[i+1:] {
			if pdoBoundMayRebind(ctx, semanticquery.CallArgument(binding.Args, 0, "param"), semanticquery.CallArgument(later.Args, 0, "param")) {
				rebound = true
			}
		}
		if rebound {
			continue
		}
		for _, event := range semanticquery.ScopeCallEvents(ctx, c) {
			a, ok := event.(*syntax.Assign)
			if !ok || a.Span().Start <= binding.Span().End || a.Span().End >= c.Span().Start || !semanticquery.NativeDominates(a, c) {
				continue
			}
			v, ok := a.Var.(*syntax.Variable)
			if ok && v.Name == variable.Name && !a.ByRef {
				ctx.ReportNode(c, message)
				return
			}
		}
	}
}

// pdoBoundMayRebind reports whether a later binding can target the placeholder
// of an earlier one. PDO accepts a named placeholder with or without its
// leading colon; a name or position that is not statically known may match.
func pdoBoundMayRebind(ctx *analysis.Context, a, b syntax.Expr) bool {
	an, as := semanticquery.NativeString(ctx, a)
	bn, bs := semanticquery.NativeString(ctx, b)
	if as && bs {
		return strings.TrimPrefix(an, ":") == strings.TrimPrefix(bn, ":")
	}
	ai, ak := semanticquery.NativeInt(ctx, a)
	bi, bk := semanticquery.NativeInt(ctx, b)
	if ak && bk {
		return ai == bi
	}
	return (!as && !ak) || (!bs && !bk)
}
