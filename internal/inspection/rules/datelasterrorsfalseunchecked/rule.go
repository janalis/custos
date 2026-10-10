// Package datelasterrorsfalseunchecked implements the DateLastErrorsFalseUnchecked inspection.
package datelasterrorsfalseunchecked

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

type rule struct{}

func New() analysis.Rule              { return rule{} }
func (rule) ID() string               { return "DateLastErrorsFalseUnchecked" }
func (rule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KArrayDimFetch} }

const message = "Handle the false result before reading parsing diagnostics."

func (rule) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP82 {
		return
	}
	fetch := n.(*syntax.ArrayDimFetch)
	origin := semanticquery.NativeValue(ctx, fetch.Var)
	if ctx.Flow().Excludes(fetch.Var, "false") {
		return
	}
	if semanticquery.NativeDateStatic(ctx, origin, "getLastErrors") == nil {
		return
	}
	for p := fetch.Parent(); p != nil; p = p.Parent() {
		conditional, ok := p.(*syntax.If)
		if !ok {
			continue
		}
		if conditional.Body.Span().Contains(fetch.Span()) && astquery.Equivalent(ctx.File, syntax.UnwrapParens(conditional.Cond), fetch.Var) {
			return
		}
	}
	ctx.ReportNode(fetch, message)
}
func (rule) Semantic() {}
