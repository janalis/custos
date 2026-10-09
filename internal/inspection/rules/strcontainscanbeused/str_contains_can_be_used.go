package strcontainscanbeused

import (
	"strings"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/semanticquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// strContainsCanBeUsed reports `strpos($h, $n) !== false` checks that
// str_contains() expresses directly.
type strContainsCanBeUsed struct{}

// Semantic marks the rule as needing the project index (user functions with
// the matched names may be declared in other files).
func (strContainsCanBeUsed) Semantic()  {}
func (strContainsCanBeUsed) ID() string { return "StrContainsCanBeUsed" }
func (strContainsCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KFuncCall}
}

func (strContainsCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP80 { // E1
		return
	}
	call := n.(*syntax.FuncCall)
	m, ok := semanticquery.MatchStrCallCompare(ctx, call, 2, "strpos", "mb_strpos") // D1-D3
	if !ok {
		return
	}
	if c, ok := m.Other.(*syntax.ConstFetch); !ok || c.Name == nil || !strings.EqualFold(strings.TrimPrefix(c.Name.Value, `\`), "false") { // D4
		return
	}
	semanticquery.ReportStrCallReplacement(ctx, m, "str_contains", m.Args[1], m.Cmp.Op.Kind == syntax.TIsIdentical) // D5
}
