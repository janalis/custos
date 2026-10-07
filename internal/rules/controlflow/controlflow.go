// Package controlflow holds the rules of the "Control flow" group.
package controlflow

import (
	"sort"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

var registered []analysis.Rule

// register adds a rule to the group; call it from an init() in the rule's file.
func register(r analysis.Rule) { registered = append(registered, r) }

// Rules returns the group's rules sorted by ID.
func Rules() []analysis.Rule {
	out := append([]analysis.Rule(nil), registered...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}

// keywordSpan returns the span of n's first token (a statement's leading
// keyword). Statement nodes always start at a significant token.
func keywordSpan(ctx *analysis.Context, n syntax.Node) syntax.Span {
	t, _ := util.NextSignificant(ctx.File, n.Span().Start)
	return syntax.Span{Start: t.Start, End: t.End}
}
