package misorderedmodifiers

import (
	"slices"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// misorderedModifiers reports method modifiers not written in the order
// final/abstract, visibility, static.
type misorderedModifiers struct{}

func (misorderedModifiers) ID() string               { return "MisorderedModifiers" }
func (misorderedModifiers) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethod} }

// mmRank orders modifier keywords; visibility keywords (and anything else the
// parser accepts as a modifier) keep the default rank 0 and their source order.
var mmRank = map[syntax.TokenKind]int{syntax.TFinal: -2, syntax.TAbstract: -1, syntax.TStatic: 1}

func (misorderedModifiers) Check(ctx *analysis.Context, n syntax.Node) {
	m := n.(*syntax.Method)
	mods := m.Modifiers
	if len(mods) < 2 || !(mods.Has(syntax.TStatic) || mods.Has(syntax.TAbstract) || mods.Has(syntax.TFinal)) { // D1, D2, E1, E2
		return
	}
	sorted := slices.Clone(mods) // D3: only real keyword tokens, never comment text
	slices.SortStableFunc(sorted, func(a, b syntax.TokenRef) int { return mmRank[a.Kind] - mmRank[b.Kind] })
	if slices.Equal(sorted, mods) { // D4, E3
		return
	}
	words := make([]string, len(sorted))
	for i, t := range sorted {
		words[i] = strings.ToLower(string(ctx.Src[t.Span.Start:t.Span.End]))
	}
	span := syntax.Span{Start: mods[0].Span.Start, End: mods[len(mods)-1].Span.End}
	ctx.Report(span, "Reorder modifiers as: "+strings.Join(words, " ")+".", diagnostic.Fix{
		Title: "Reorder modifiers",
		Edits: func() []diagnostic.TextEdit {
			gaps := make([]string, len(mods)-1)
			commented := false
			for i := range gaps {
				gaps[i] = string(ctx.Src[mods[i].Span.End:mods[i+1].Span.Start])
				commented = commented || strings.TrimSpace(gaps[i]) != ""
			}
			var b strings.Builder
			for i, w := range words {
				b.WriteString(w)
				if i < len(gaps) {
					if commented { // F2: keep comments and their layout in place
						b.WriteString(gaps[i])
					} else { // F1
						b.WriteByte(' ')
					}
				}
			}
			return []diagnostic.TextEdit{{Span: span, NewText: b.String()}}
		},
	})
}
