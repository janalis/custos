package sodiumkdfcontextlengthmismatch

import (
	"os"
	"path/filepath"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/testing/conformance"
)

func TestFixtures(t *testing.T) {
	for _, name := range []string{"basic", "negative"} {
		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("../../../../testdata/rules", "SodiumKdfContextLengthMismatch", name+".php"))
			if err != nil {
				t.Fatal(err)
			}
			src, want, err := conformance.ParseMarkup(data)
			if err != nil {
				t.Fatal(err)
			}
			engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SodiumKdfContextLengthMismatch"}})
			if err != nil {
				t.Fatal(err)
			}
			got := engine.Analyze(syntax.Parse("case.php", src, syntax.Options{}))
			if len(got) != len(want) {
				t.Fatalf("got %d want %d: %+v", len(got), len(want), got)
			}
			for i, g := range got {
				w := want[i]
				if int(g.Span.Start) != w.Start || int(g.Span.End) != w.End || g.Message != w.Message || g.Severity != w.Severity {
					t.Fatalf("got %+v want %+v", g, w)
				}
				if len(g.Fixes) != 0 {
					t.Fatal("unexpected fix")
				}
			}
		})
	}
}
