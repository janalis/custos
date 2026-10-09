package astquery

import (
	"strings"

	"custos/internal/php/syntax"
)

// KeywordSpan returns the span of n's first token (a statement's leading
// keyword). Statement nodes always start at a significant token.
func KeywordSpan(file *syntax.File, n syntax.Node) syntax.Span {
	t, _ := NextSignificant(file, n.Span().Start)
	return syntax.Span{Start: t.Start, End: t.End}
}

// BracedBlock returns s as a `{ … }` block.
func BracedBlock(src []byte, s syntax.Stmt) (*syntax.Block, bool) {
	b, ok := s.(*syntax.Block)
	if !ok || b.Alt || b.Span().Len() < 2 || src[b.Span().Start] != '{' {
		return nil, false
	}
	return b, true
}

// WrittenArguments returns the call's arguments as written (spreads and named
// arguments included); nil for a first-class callable `f(...)`.
func WrittenArguments(list *syntax.ArgList) []*syntax.Arg {
	out := make([]*syntax.Arg, 0, len(list.Args))
	for _, a := range list.Args {
		arg, ok := a.(*syntax.Arg)
		if !ok || arg.Value == nil {
			return nil
		}
		out = append(out, arg)
	}
	return out
}

// ArgumentNodes returns the argument nodes of a call's argument list (nil list ok).
func ArgumentNodes(list *syntax.ArgList) []syntax.Expr {
	if list == nil {
		return nil
	}
	return list.Args
}

// NeedsParensAsComparison reports whether replacing n with an unparenthesised
// `a == b` would change the meaning (spec Divergences): n is the operand of
// an operator binding at least as tight as `==`.
func NeedsParensAsComparison(n syntax.Node) bool {
	switch p := n.Parent().(type) {
	case *syntax.Binary:
		switch p.Op.Kind {
		case syntax.TBooleanAnd, syntax.TBooleanOr, syntax.TAnd, syntax.TOr, syntax.TXor,
			syntax.TCoalesce, syntax.TAmpersand, syntax.TBar, syntax.TCaret:
			return false
		}
		return true
	case *syntax.Unary, *syntax.Instanceof:
		return true
	}
	return false
}

// MethodStatements returns the body's statements, ignoring empty ones.
func MethodStatements(b *syntax.Block) []syntax.Stmt {
	var out []syntax.Stmt
	for _, s := range b.Stmts {
		if s.Span().Len() > 0 {
			out = append(out, s)
		}
	}
	return out
}

// SquashWhitespace removes all whitespace (text-level equivalence fallback).
func SquashWhitespace(s string) string {
	return strings.Join(strings.Fields(s), "")
}
