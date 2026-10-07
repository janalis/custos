package fix

import (
	"testing"

	"custos/internal/analysis"
	"custos/internal/syntax"
)

func TestApply(t *testing.T) {
	src := []byte("0123456789")
	out, n := Apply(src, []analysis.TextEdit{
		{Span: syntax.Span{Start: 8, End: 9}, NewText: "X"},
		{Span: syntax.Span{Start: 1, End: 3}, NewText: "ab"},
		{Span: syntax.Span{Start: 2, End: 4}, NewText: "overlap"}, // dropped
		{Span: syntax.Span{Start: 5, End: 5}, NewText: "+"},
	})
	if string(out) != "0ab34+567X9" || n != 3 {
		t.Fatalf("got %q (%d)", out, n)
	}
}
