package util

import (
	"testing"

	"custos/internal/phpver"
	"custos/internal/syntax"
)

func parse(t *testing.T, src string) *syntax.File {
	t.Helper()
	f := syntax.Parse("t.php", []byte(src), syntax.Options{Version: phpver.Max})
	if len(f.Errors) > 0 {
		t.Fatalf("parse %q: %v", src, f.Errors)
	}
	return f
}

// firstExpr returns the expression of the first ExprStmt.
func firstExpr(t *testing.T, f *syntax.File) syntax.Expr {
	t.Helper()
	for _, s := range f.Stmts {
		if es, ok := s.(*syntax.ExprStmt); ok {
			return es.Expr
		}
	}
	t.Fatal("no expression statement")
	return nil
}

func text(f *syntax.File, n syntax.Node) string {
	s := n.Span()
	return string(f.Src[s.Start:s.End])
}

func spanText(f *syntax.File, s syntax.Span) string { return string(f.Src[s.Start:s.End]) }

func TestUnwrapParens(t *testing.T) {
	f := parse(t, "<?php ((($a + 1)));")
	e := firstExpr(t, f)
	if got := text(f, syntax.UnwrapParens(e)); got != "$a + 1" {
		t.Fatalf("got %q", got)
	}
	f = parse(t, "<?php $a;")
	if got := text(f, syntax.UnwrapParens(firstExpr(t, f))); got != "$a" {
		t.Fatalf("got %q", got)
	}
}

func TestParentSkipParens(t *testing.T) {
	f := parse(t, "<?php !(($a));")
	not := firstExpr(t, f).(*syntax.Unary)
	v := syntax.UnwrapParens(not.Expr)
	parent, child := ParentSkipParens(v)
	if parent != syntax.Node(not) {
		t.Fatalf("parent = %T", parent)
	}
	if child != syntax.Node(not.Expr) {
		t.Fatalf("child = %s", text(f, child))
	}
	parent, child = ParentSkipParens(not)
	if _, ok := parent.(*syntax.ExprStmt); !ok || child != syntax.Node(not) {
		t.Fatalf("parent = %T", parent)
	}
}

func TestTokenQueries(t *testing.T) {
	src := "<?php foo( 1 ); /* c */ bar();"
	f := parse(t, src)
	off := uint32(len("<?php foo("))
	if tok, ok := TokenBefore(f, off); !ok || tok.Kind != syntax.TLParen {
		t.Fatalf("TokenBefore = %v %v", tok, ok)
	}
	if tok, ok := TokenAfter(f, off); !ok || tok.Kind != syntax.TWhitespace {
		t.Fatalf("TokenAfter = %v %v", tok, ok)
	}
	if _, ok := TokenAfter(f, uint32(len("<?php f"))); ok {
		t.Fatal("TokenAfter inside a token must fail")
	}
	if tok, ok := NextSignificant(f, off); !ok || tok.Kind != syntax.TLNumber {
		t.Fatalf("NextSignificant = %v", tok)
	}
	if _, ok := NextSignificant(f, uint32(len(src))); ok {
		t.Fatal("NextSignificant at EOF must fail")
	}
	whole := syntax.Span{Start: 0, End: uint32(len(src))}
	if tok, ok := FindToken(f, whole, syntax.TSemicolon); !ok || spanText(f, syntax.Span{Start: tok.Start, End: tok.End}) != ";" || tok.Start != uint32(len("<?php foo( 1 )")) {
		t.Fatalf("FindToken = %v", tok)
	}
	if _, ok := FindToken(f, syntax.Span{Start: 0, End: 9}, syntax.TSemicolon); ok {
		t.Fatal("FindToken outside span")
	}
	if !HasComment(f, whole) {
		t.Fatal("HasComment = false")
	}
	if HasComment(f, syntax.Span{Start: 0, End: uint32(len("<?php foo( 1 );"))}) {
		t.Fatal("HasComment = true before the comment")
	}
}

func TestWhitespaceExtension(t *testing.T) {
	src := "<?php a();\n  ; // x\n;/* y */\n ;"
	f := parse(t, src)
	semi := func(nth int) syntax.Span {
		n := 0
		for _, tok := range f.Tokens {
			if tok.Kind == syntax.TSemicolon {
				if n == nth {
					return syntax.Span{Start: tok.Start, End: tok.End}
				}
				n++
			}
		}
		t.Fatal("no such semicolon")
		return syntax.Span{}
	}
	if got := spanText(f, WithLeadingWhitespace(f, semi(1))); got != "\n  ;" {
		t.Fatalf("leading ws: %q", got)
	}
	if got := spanText(f, WithLeadingWhitespace(f, semi(2))); got != ";" {
		t.Fatalf("after line comment: %q", got)
	}
	if got := spanText(f, WithLeadingWhitespace(f, semi(3))); got != "\n ;" {
		t.Fatalf("after block comment: %q", got)
	}
	if got := spanText(f, WithLeadingWhitespace(f, semi(0))); got != ";" {
		t.Fatalf("no ws: %q", got)
	}
	if got := spanText(f, WithTrailingWhitespace(f, semi(0))); got != ";\n  " {
		t.Fatalf("trailing ws: %q", got)
	}
}

func TestNeedsParensAsUnaryOperand(t *testing.T) {
	for src, want := range map[string]bool{
		"<?php $a && $b;":        true,
		"<?php $a instanceof B;": true,
		"<?php $a ? $b : $c;":    true,
		"<?php $a = 1;":          true,
		"<?php $a ?? $b;":        true,
		"<?php f($a);":           false,
		"<?php $a;":              false,
		"<?php -$a;":             false,
		"<?php ($a + $b);":       false,
		"<?php fn() => 1;":       true,
		"<?php $a->b[1];":        false,
	} {
		f := parse(t, src)
		if got := NeedsParensAsUnaryOperand(firstExpr(t, f)); got != want {
			t.Errorf("%s: got %v", src, got)
		}
	}
}

func TestTokenEdges(t *testing.T) {
	f := parse(t, "<?php $a;")
	if _, ok := TokenBefore(f, 0); ok {
		t.Fatal("nothing before offset 0")
	}
	if _, ok := TokenBefore(f, 7); ok {
		t.Fatal("offset inside a token")
	}
	sp := syntax.Span{Start: 6, End: 8}
	if got := WithTrailingWhitespace(f, sp); got != sp {
		t.Fatalf("no whitespace to add: %v", got)
	}
}
