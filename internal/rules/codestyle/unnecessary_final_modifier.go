package codestyle

import (
	"strings"

	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// unnecessaryFinalModifier reports `final` on methods that cannot be
// overridden anyway: methods of a final class and private non-magic methods.
type unnecessaryFinalModifier struct{}

func init() { register(unnecessaryFinalModifier{}) }

func (unnecessaryFinalModifier) ID() string { return "UnnecessaryFinalModifier" }

func (unnecessaryFinalModifier) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KMethod} }

func (unnecessaryFinalModifier) Check(ctx *analysis.Context, n syntax.Node) {
	m := n.(*syntax.Method) // D2: property hooks are a different node kind
	var final syntax.Span
	found := false
	for _, t := range m.Modifiers { // D1
		if t.Kind == syntax.TFinal {
			final, found = t.Span, true
			break
		}
	}
	if !found {
		return
	}
	cls, _ := m.Parent().(*syntax.ClassLike)
	redundant := cls != nil && cls.Modifiers.Has(syntax.TFinal) || // D3a
		m.Modifiers.Has(syntax.TPrivate) && !strings.HasPrefix(m.Name.Value, "__") // D3b
	if !redundant {
		return
	}
	f := ctx.File
	ctx.ReportNode(m.Name, "Redundant final: the method cannot be overridden anyway.", analysis.Fix{
		Title: "Remove 'final'",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: util.WithTrailingWhitespace(f, final)}}
		},
	})
}
