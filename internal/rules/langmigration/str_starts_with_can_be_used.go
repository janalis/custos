package langmigration

import (
	"custos/internal/analysis"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// strStartsWithCanBeUsed reports `strpos($h, $n) === 0` checks that
// str_starts_with() expresses directly.
type strStartsWithCanBeUsed struct{}

func init() { register(strStartsWithCanBeUsed{}) }

// Semantic marks the rule as needing the project index (user functions with
// the matched names may be declared in other files).
func (strStartsWithCanBeUsed) Semantic() {}

func (strStartsWithCanBeUsed) ID() string { return "StrStartsWithCanBeUsed" }

func (strStartsWithCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (strStartsWithCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP80 { // E1
		return
	}
	m, ok := matchStrCallCompare(ctx, n.(*syntax.FuncCall), 2, "strpos", "mb_strpos") // D1-D3
	if !ok {
		return
	}
	if lit, ok := m.other.(*syntax.Literal); !ok || lit.LitKind != syntax.LitInt || lit.Raw != "0" { // D4
		return
	}
	reportStrCallReplacement(ctx, m, "str_starts_with", m.args[1], m.cmp.Op.Kind == syntax.TIsNotIdentical) // D5
}
