package fixing

import (
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestApply(t *testing.T) {
	src := []byte("0123456789")
	out, n := Apply(src, []diagnostic.TextEdit{
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
	ctx.ReportNode(n, "x", diagnostic.Fix{Title: "bad", Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{{Span: syntax.Span{Start: 1 << 20, End: 1 << 20}, NewText: "y"}}
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

func TestOverlapsAnyInsertions(t *testing.T) {
	ins := func(p uint32) diagnostic.TextEdit {
		return diagnostic.TextEdit{Span: syntax.Span{Start: p, End: p}, NewText: "\\"}
	}
	rep := func(s, e uint32) diagnostic.TextEdit {
		return diagnostic.TextEdit{Span: syntax.Span{Start: s, End: e}, NewText: "r"}
	}
	for _, tc := range []struct {
		a, b diagnostic.TextEdit
		want bool
	}{
		{ins(5), rep(5, 9), true},  // `\` before a call another fix rewrites
		{rep(5, 9), ins(5), true},  // same, other order
		{ins(9), rep(5, 9), true},  // insertion at the end of a replacement
		{rep(5, 9), ins(7), true},  // insertion inside
		{ins(4), rep(5, 9), false}, // disjoint
		{ins(5), ins(5), true},     // two insertions at one point
		{ins(5), ins(6), false},
		{rep(1, 3), rep(3, 5), false}, // adjacent replacements
	} {
		if got := overlapsAny([]diagnostic.TextEdit{tc.a}, []diagnostic.TextEdit{tc.b}); got != tc.want {
			t.Errorf("overlapsAny(%v, %v) = %v", tc.a.Span, tc.b.Span, got)
		}
	}
}

func TestApplySameStart(t *testing.T) {
	// Edits starting at one point are ordered by end: the insertion first.
	out, n := Apply([]byte("abcdef"), []diagnostic.TextEdit{
		{Span: syntax.Span{Start: 2, End: 4}, NewText: "XY"},
		{Span: syntax.Span{Start: 2, End: 2}, NewText: "+"},
	})
	if string(out) != "ab+XYef" || n != 2 {
		t.Fatalf("got %q (%d)", out, n)
	}
}
