package unused

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/syntax"
)

// unusedGotoLabel reports goto labels no goto in the same function targets.
type unusedGotoLabel struct{}

func init() { register(unusedGotoLabel{}) }

func (unusedGotoLabel) ID() string { return "UnusedGotoLabel" }

func (unusedGotoLabel) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KLabel} }

func (unusedGotoLabel) Check(ctx *analysis.Context, n syntax.Node) {
	l := n.(*syntax.Label)
	if l.Name == nil || l.Span().Len() == 0 {
		return
	}
	fn := util.EnclosingFuncLike(l) // D1
	if fn == nil {
		return
	}
	body := util.FuncLikeBody(fn) // D2
	if body == nil {
		return
	}
	used := false
	syntax.Inspect(body, func(x syntax.Node) bool { // D3
		if used {
			return false
		}
		// A goto cannot leave its own function: nested function-likes and
		// classes are separate scopes (custos diverges).
		if _, ok := x.(*syntax.ClassLike); ok || (x != body && util.IsFuncLike(x)) {
			return false
		}
		if g, ok := x.(*syntax.Goto); ok && g.Label != nil && g.Label.Value == l.Name.Value {
			used = true
		}
		return !used
	})
	if used {
		return
	}
	span := l.Span()
	f := ctx.File
	ctx.Report(span, "Label '"+l.Name.Value+"' is never targeted by a goto; remove it.", analysis.Fix{
		Title: "Remove the label",
		Edits: func() []analysis.TextEdit {
			return []analysis.TextEdit{{Span: util.WithLeadingWhitespace(f, span)}}
		},
	})
}
