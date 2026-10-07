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

func TestApplyNoEdits(t *testing.T) {
	src := []byte("abc")
	if out, n := Apply(src, nil); string(out) != "abc" || n != 0 {
		t.Fatalf("got %q (%d)", out, n)
	}
}

// badFixRule offers a fix whose edit lies outside the source: Apply drops
// it, and FixSource must stop instead of looping.
type badFixRule struct{}

func (badFixRule) ID() string               { return "NestedNotOperators" }
func (badFixRule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KUnary} }
func (badFixRule) Check(ctx *analysis.Context, n syntax.Node) {
	ctx.ReportNode(n, "x", analysis.Fix{Title: "bad", Edits: func() []analysis.TextEdit {
		return []analysis.TextEdit{{Span: syntax.Span{Start: 1 << 20, End: 1 << 20}, NewText: "y"}}
	}})
}

func TestFixSourceInvalidEditStops(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{badFixRule{}}, analysis.Config{})
	if err != nil {
		t.Fatal(err)
	}
	src := []byte("<?php $a = !$b;")
	res := FixSource(e, "x.php", src, Options{})
	if string(res.Source) != string(src) || res.Applied != 0 || res.Iterations != 1 {
		t.Fatalf("got %+v", res)
	}
}
