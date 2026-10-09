package strstartswithcanbeused

import (
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// strStartsWithCanBeUsed reports `strpos($h, $n) === 0` checks that
// str_starts_with() expresses directly.
type strStartsWithCanBeUsed struct{}

// Semantic marks the rule as needing the project index (user functions with
// the matched names may be declared in other files).
func (strStartsWithCanBeUsed) Semantic()  {}
func (strStartsWithCanBeUsed) ID() string { return "StrStartsWithCanBeUsed" }
func (strStartsWithCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (strStartsWithCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP80 { // E1
		return
	}
	m, ok := semanticquery.MatchStrCallCompare(ctx, n.(*syntax.FuncCall), 2, "strpos", "mb_strpos") // D1-D3
	if !ok {
		return
	}
	if lit, ok := m.Other.(*syntax.Literal); !ok || lit.LitKind != syntax.LitInt || lit.Raw != "0" { // D4
		return
	}
	semanticquery.ReportStrCallReplacement(ctx, m, "str_starts_with", m.Args[1], m.Cmp.Op.Kind == syntax.TIsNotIdentical) // D5
}
