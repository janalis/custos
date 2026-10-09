package shortlistsyntaxcanbeused

import (
	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/astquery"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// shortListSyntaxCanBeUsed reports list(...) destructuring that can use the
// short [...] form.
type shortListSyntaxCanBeUsed struct{}

const (
	shortListAssignMsg  = "Use short destructuring syntax '[...] = ...'."
	shortListForeachMsg = "Use short destructuring syntax in foreach ('as [...]')."
)

func (shortListSyntaxCanBeUsed) ID() string { return "ShortListSyntaxCanBeUsed" }
func (shortListSyntaxCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KList}
}

func (shortListSyntaxCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpversion.PHP71 { // E1
		return
	}
	l := n.(*syntax.List)
	s := l.Span()
	if s.Len() < 6 || ctx.Src[s.End-1] != ')' {
		return // recovery node
	}
	var msg string
	switch p := l.Parent().(type) {
	case *syntax.Assign: // D1
		// Any context, statement or sub-expression (spec Divergences).
		if p.Var != syntax.Expr(l) || p.Op.Kind != syntax.TEqual {
			return
		}
		msg = shortListAssignMsg
	case *syntax.Foreach: // D2
		if p.Value != syntax.Expr(l) || (p.Key == nil && !astquery.PatternHasTarget(l.Items)) { // E4
			return
		}
		msg = shortListForeachMsg
	default: // E5: nested patterns and other contexts
		return
	}
	kw := syntax.Span{Start: s.Start, End: s.Start + 4}
	f := ctx.File
	ctx.Report(kw, msg, diagnostic.Fix{
		Title: "Use short destructuring syntax",
		Edits: func() []diagnostic.TextEdit { return shortListEdits(f, l) },
	})
}

// shortListEdits converts l and every list(...) nested in its pattern (mixing
// both forms is a compile error) to the bracket form.
func shortListEdits(f *syntax.File, l *syntax.List) []diagnostic.TextEdit {
	var edits []diagnostic.TextEdit
	syntax.Inspect(l, func(n syntax.Node) bool {
		inner, ok := n.(*syntax.List)
		if !ok {
			return true
		}
		s := inner.Span()
		lp, ok := astquery.FindToken(f, s, syntax.TLParen)
		if !ok || f.Src[s.End-1] != ')' {
			return true
		}
		edits = append(edits,
			diagnostic.TextEdit{Span: syntax.Span{Start: s.Start, End: lp.Start}},
			diagnostic.TextEdit{Span: syntax.Span{Start: lp.Start, End: lp.End}, NewText: "["},
			diagnostic.TextEdit{Span: syntax.Span{Start: s.End - 1, End: s.End}, NewText: "]"},
		)
		return true
	})
	return edits
}
