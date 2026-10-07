// Package util holds small, allocation-free helpers shared by rules: AST
// navigation through parentheses and token-stream queries on byte spans.
package util

import "custos/internal/syntax"

// UnwrapParens strips any number of enclosing parentheses from e.
func UnwrapParens(e syntax.Expr) syntax.Expr {
	for {
		p, ok := e.(*syntax.Paren)
		if !ok || p.Expr == nil {
			return e
		}
		e = p.Expr
	}
}

// ParentSkipParens returns the first ancestor of n that is not a Paren, and
// the outermost Paren (or n itself) directly below it.
func ParentSkipParens(n syntax.Node) (parent, child syntax.Node) {
	child = n
	parent = n.Parent()
	for parent != nil {
		if _, ok := parent.(*syntax.Paren); !ok {
			break
		}
		child = parent
		parent = parent.Parent()
	}
	return parent, child
}

// TokenIndex returns the index of the first token of f starting at or after
// off (len(f.Tokens) when none does).
func TokenIndex(f *syntax.File, off uint32) int {
	toks := f.Tokens
	lo, hi := 0, len(toks)
	for lo < hi {
		m := int(uint(lo+hi) >> 1)
		if toks[m].Start < off {
			lo = m + 1
		} else {
			hi = m
		}
	}
	return lo
}

// TokenBefore returns the token ending exactly at off (the token stream is
// contiguous, so this is the token immediately preceding off).
func TokenBefore(f *syntax.File, off uint32) (syntax.Token, bool) {
	i := TokenIndex(f, off) - 1
	if i < 0 || f.Tokens[i].End != off {
		return syntax.Token{}, false
	}
	return f.Tokens[i], true
}

// TokenAfter returns the token starting exactly at off.
func TokenAfter(f *syntax.File, off uint32) (syntax.Token, bool) {
	i := TokenIndex(f, off)
	if i >= len(f.Tokens) || f.Tokens[i].Start != off {
		return syntax.Token{}, false
	}
	return f.Tokens[i], true
}

// NextSignificant returns the first non-trivia token starting at or after
// off; ok is false at end of file.
func NextSignificant(f *syntax.File, off uint32) (syntax.Token, bool) {
	for i := TokenIndex(f, off); i < len(f.Tokens); i++ {
		if t := f.Tokens[i]; !t.Kind.IsTrivia() {
			return t, t.Kind != syntax.TEOF
		}
	}
	return syntax.Token{}, false
}

// FindToken returns the first token of kind k lying entirely inside span.
func FindToken(f *syntax.File, span syntax.Span, k syntax.TokenKind) (syntax.Token, bool) {
	for i := TokenIndex(f, span.Start); i < len(f.Tokens); i++ {
		t := f.Tokens[i]
		if t.End > span.End {
			break
		}
		if t.Kind == k {
			return t, true
		}
	}
	return syntax.Token{}, false
}

// HasComment reports whether a comment or doc comment lies inside span.
func HasComment(f *syntax.File, span syntax.Span) bool {
	for i := TokenIndex(f, span.Start); i < len(f.Tokens); i++ {
		t := f.Tokens[i]
		if t.End > span.End {
			break
		}
		if t.Kind == syntax.TComment || t.Kind == syntax.TDocComment {
			return true
		}
	}
	return false
}

// WithLeadingWhitespace extends span to the left over the whitespace token
// immediately preceding it, if any. The whitespace is kept when removing it
// would glue the code onto a preceding line comment (`// x` + newline).
func WithLeadingWhitespace(f *syntax.File, span syntax.Span) syntax.Span {
	ws, ok := TokenBefore(f, span.Start)
	if !ok || ws.Kind != syntax.TWhitespace {
		return span
	}
	if prev, ok := TokenBefore(f, ws.Start); ok && isLineComment(f, prev) {
		return span
	}
	return syntax.Span{Start: ws.Start, End: span.End}
}

// WithTrailingWhitespace extends span to the right over the whitespace token
// immediately following it, if any.
func WithTrailingWhitespace(f *syntax.File, span syntax.Span) syntax.Span {
	if ws, ok := TokenAfter(f, span.End); ok && ws.Kind == syntax.TWhitespace {
		return syntax.Span{Start: span.Start, End: ws.End}
	}
	return span
}

func isLineComment(f *syntax.File, t syntax.Token) bool {
	return t.Kind == syntax.TComment && !(t.End-t.Start >= 2 && f.Src[t.Start] == '/' && f.Src[t.Start+1] == '*')
}

// NeedsParensAsUnaryOperand reports whether e must be parenthesised when it
// becomes the operand of a prefix operator or cast (`!`, `(bool)`, `-`, …):
// binary operations, instanceof, ternaries, assignments and the
// low-precedence keyword expressions.
func NeedsParensAsUnaryOperand(e syntax.Expr) bool {
	switch e.(type) {
	case *syntax.Binary, *syntax.Instanceof, *syntax.Ternary, *syntax.Assign,
		*syntax.Print, *syntax.Yield, *syntax.YieldFrom, *syntax.Include,
		*syntax.Throw, *syntax.ArrowFunction:
		return true
	}
	return false
}
