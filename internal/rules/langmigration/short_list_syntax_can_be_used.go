package langmigration

import (
	"custos/internal/analysis"
	"custos/internal/analysis/util"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// shortListSyntaxCanBeUsed reports list(...) destructuring that can use the
// short [...] form.
type shortListSyntaxCanBeUsed struct{}

func init() { register(shortListSyntaxCanBeUsed{}) }

const (
	shortListAssignMsg  = "Use short destructuring syntax '[...] = ...'."
	shortListForeachMsg = "Use short destructuring syntax in foreach ('as [...]')."
)

func (shortListSyntaxCanBeUsed) ID() string { return "ShortListSyntaxCanBeUsed" }

func (shortListSyntaxCanBeUsed) Kinds() []syntax.NodeKind {
	return []syntax.NodeKind{syntax.KList}
}

func (shortListSyntaxCanBeUsed) Check(ctx *analysis.Context, n syntax.Node) {
	if ctx.PHP < phpver.PHP71 { // E1
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
		if p.Value != syntax.Expr(l) || (p.Key == nil && !patternHasTarget(l.Items)) { // E4
			return
		}
		msg = shortListForeachMsg
	default: // E5: nested patterns and other contexts
		return
	}
	kw := syntax.Span{Start: s.Start, End: s.Start + 4}
	f := ctx.File
	ctx.Report(kw, msg, analysis.Fix{
		Title: "Use short destructuring syntax",
		Edits: func() []analysis.TextEdit { return shortListEdits(f, l) },
	})
}

// shortListEdits converts l and every list(...) nested in its pattern (mixing
// both forms is a compile error) to the bracket form.
func shortListEdits(f *syntax.File, l *syntax.List) []analysis.TextEdit {
	var edits []analysis.TextEdit
	syntax.Inspect(l, func(n syntax.Node) bool {
		inner, ok := n.(*syntax.List)
		if !ok {
			return true
		}
		s := inner.Span()
		lp, ok := util.FindToken(f, s, syntax.TLParen)
		if !ok || f.Src[s.End-1] != ')' {
			return true
		}
		edits = append(edits,
			analysis.TextEdit{Span: syntax.Span{Start: s.Start, End: lp.Start}},
			analysis.TextEdit{Span: syntax.Span{Start: lp.Start, End: lp.End}, NewText: "["},
			analysis.TextEdit{Span: syntax.Span{Start: s.End - 1, End: s.End}, NewText: "]"},
		)
		return true
	})
	return edits
}

// patternHasTarget reports whether a destructuring pattern assigns anything
// (recursively through nested patterns).
func patternHasTarget(items []*syntax.ArrayItem) bool {
	for _, it := range items {
		if it == nil || it.Value == nil {
			continue
		}
		switch v := it.Value.(type) {
		case *syntax.List:
			if patternHasTarget(v.Items) {
				return true
			}
		case *syntax.Array:
			if patternHasTarget(v.Items) {
				return true
			}
		default:
			return true
		}
	}
	return false
}
