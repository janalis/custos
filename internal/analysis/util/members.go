package util

import "custos/internal/syntax"

// MemberRemovalSpans returns the byte ranges to delete to remove a class
// member declaration n: its own span plus the whitespace run right after it
// and, when the previous significant token (skipping whitespace and
// ordinary comments) is a doc comment, that doc comment. The whitespace
// before the member stays, so it now precedes the next member.
func MemberRemovalSpans(f *syntax.File, n syntax.Node) []syntax.Span {
	span := WithTrailingWhitespace(f, n.Span())
	out := []syntax.Span{span}
	for i := TokenIndex(f, n.Span().Start) - 1; i >= 0; i-- {
		t := f.Tokens[i]
		if t.Kind == syntax.TWhitespace || t.Kind == syntax.TComment {
			continue
		}
		if t.Kind == syntax.TDocComment && !n.Span().Contains(syntax.Span{Start: t.Start, End: t.End}) {
			out = append([]syntax.Span{{Start: t.Start, End: t.End}}, out...)
		}
		break
	}
	return out
}
