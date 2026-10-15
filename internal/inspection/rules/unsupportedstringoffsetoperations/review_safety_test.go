package unsupportedstringoffsetoperations

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestVariadicReferenceAfterWrite(t *testing.T) {
	for _, tc := range []struct {
		name, reference string
		findings        int
	}{{"reference", "&", 0}, {"value", "", 1}} {
		t.Run(tc.name, func(t *testing.T) {
			e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{PHP: phpversion.PHP85, EnableAll: true})
			if err != nil {
				t.Fatal(err)
			}
			src := `<?php function receive(` + tc.reference + `...$values) {} function append(string $s,$other) { $s[]='x'; receive($other,$s); }`
			f := syntax.ParseBest("variadic.php", []byte(src), syntax.Options{Version: phpversion.PHP85})
			if len(f.Errors) != 0 {
				t.Fatal(f.Errors)
			}
			if got := e.Analyze(f); len(got) != tc.findings {
				t.Fatalf("findings: %+v", got)
			}
		})
	}
}
