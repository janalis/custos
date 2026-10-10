package sortcomparatorreturnsboolean

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/testing/conformance"
)

func TestNativeFixtureRangesAndFixes(t *testing.T) {
	for _, name := range []string{"basic", "negative"} {
		t.Run(name, func(t *testing.T) {
			base := filepath.Join("../../../../testdata/rules", "SortComparatorReturnsBoolean", name)
			marked, err := os.ReadFile(base + ".php")
			if err != nil {
				t.Fatal(err)
			}
			src, want, err := conformance.ParseMarkup(marked)
			if err != nil {
				t.Fatal(err)
			}
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SortComparatorReturnsBoolean"}})
			if err != nil {
				t.Fatal(err)
			}
			got := e.Analyze(syntax.Parse("test.php", src, syntax.Options{}))
			if len(got) != len(want) {
				t.Fatalf("got %d findings, want %d: %+v", len(got), len(want), got)
			}
			for i, g := range got {
				w := want[i]
				if int(g.Span.Start) != w.Start || int(g.Span.End) != w.End || g.Message != w.Message || g.Severity != w.Severity {
					t.Fatalf("finding mismatch: got %+v, want %+v", g, w)
				}
			}
			expected, err := os.ReadFile(base + ".fixed.php")
			if os.IsNotExist(err) {
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var edits []diagnostic.TextEdit
			for _, g := range got {
				if len(g.Fixes) > 0 {
					edits = append(edits, g.Fixes[0].Edits()...)
				}
			}
			fixed, _ := fixing.Apply(src, edits)
			if !bytes.Equal(fixed, expected) {
				t.Fatalf("fixed output mismatch:\ngot %s\nwant %s", fixed, expected)
			}
			if again := e.Analyze(syntax.Parse("fixed.php", fixed, syntax.Options{})); len(again) != 0 {
				t.Fatalf("fix did not remove diagnostic: %+v", again)
			}
		})
	}
}
