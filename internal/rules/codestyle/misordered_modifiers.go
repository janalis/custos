package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// misorderedModifiers reports method modifiers not written in the order
// final/abstract, visibility, static.
type misorderedModifiers struct{}

func init() { register(misorderedModifiers{}) }

func (misorderedModifiers) ID() string { return "MisorderedModifiers" }

func (misorderedModifiers) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethod} }

var mmCanonical = [...]string{"final", "abstract", "public", "protected", "private", "static"}

func (misorderedModifiers) Check(ctx *analysis.Context, n syntax.Node) {
	m := n.(*syntax.Method)
	mods := m.Modifiers
	if len(mods) == 0 || !(mods.Has(syntax.TStatic) || mods.Has(syntax.TAbstract) || mods.Has(syntax.TFinal)) { // D1, E1
		return
	}
	span := syntax.Span{Start: mods[0].Span.Start, End: mods[len(mods)-1].Span.End}
	f := ctx.File
	var parts []string
	for i := util.TokenIndex(f, span.Start); i < len(f.Tokens); i++ { // D2, D3
		t := f.Tokens[i]
		if t.End > span.End {
			break
		}
		if t.Kind != syntax.TWhitespace {
			parts = append(parts, strings.ToLower(string(ctx.Src[t.Start:t.End])))
		}
	}
	if len(parts) < 2 { // E2
		return
	}
	original := strings.Join(parts, " ")
	kept := make([]string, 0, len(mmCanonical))
	for _, k := range mmCanonical {
		if strings.Contains(original, k) {
			kept = append(kept, k)
		}
	}
	expected := strings.Join(kept, " ")
	if original == expected { // D4, E3
		return
	}
	ctx.Report(span, "Reorder modifiers as: "+expected+".", analysis.Fix{
		Title: "Reorder modifiers",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: span, NewText: expected}}
		},
	})
}
