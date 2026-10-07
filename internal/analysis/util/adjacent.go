package util

import "custos/internal/syntax"

// PrevStmtNoDoc returns the statement preceding s in its statement list,
// provided no doc comment (`/** */`) lies between the two; ordinary comments
// are skipped.
func PrevStmtNoDoc(f *syntax.File, s syntax.Stmt) (syntax.Stmt, bool) {
	prev, ok := PrevStmt(f, s)
	if !ok {
		return nil, false
	}
	if DocCommentBetween(f, prev.Span().End, s.Span().Start) {
		return nil, false
	}
	return prev, true
}

// DocCommentBetween reports whether a doc comment token lies in [start, end).
func DocCommentBetween(f *syntax.File, start, end uint32) bool {
	for i := TokenIndex(f, start); i < len(f.Tokens); i++ {
		t := f.Tokens[i]
		if t.End > end {
			break
		}
		if t.Kind == syntax.TDocComment {
			return true
		}
	}
	return false
}

// DocCommentBefore returns the doc comment token immediately preceding off
// (only whitespace in between).
func DocCommentBefore(f *syntax.File, off uint32) (syntax.Token, bool) {
	for i := TokenIndex(f, off) - 1; i >= 0; i-- {
		t := f.Tokens[i]
		switch t.Kind {
		case syntax.TWhitespace:
			continue
		case syntax.TDocComment:
			return t, true
		}
		break
	}
	return syntax.Token{}, false
}
