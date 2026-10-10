package bytepaddingusedforunicodewidth_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/bytepaddingusedforunicodewidth"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$padded=str_pad('ñ',2,'.');\n", 1},
		{"negative0", "<?php\nstr_pad('abc',10);\n", 0},
		{"negative1", "<?php\nstr_pad($x,10);\n", 0},
		{"negative2", "<?php\nstr_pad('',10);\n", 0},
		{"negative3", "<?php\ntrim('x');\n", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rule.New().(analysis.SemanticRule).Semantic()
			e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{Only: []string{rule.New().ID()}})
			if err != nil {
				t.Fatal(err)
			}
			f := syntax.Parse("test.php", []byte(tc.source), syntax.Options{})
			findings := e.Analyze(f)
			if len(findings) != tc.count {
				t.Fatalf("got %d findings, want %d: %+v", len(findings), tc.count, findings)
			}
		})
	}
}
