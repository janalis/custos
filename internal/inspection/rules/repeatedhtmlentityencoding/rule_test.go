package repeatedhtmlentityencoding_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/repeatedhtmlentityencoding"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\necho htmlspecialchars(htmlspecialchars($text));\n", 1},
		{"negative0", "<?php\nhtmlspecialchars($s);\n", 0},
		{"negative1", "<?php\nhtmlspecialchars(htmlspecialchars($s),double_encode:false);\n", 0},
		{"negative2", "<?php\nhtmlspecialchars(htmlentities($s));\n", 0},
		{"negative3", "<?php\nhtmlspecialchars(htmlspecialchars($s),ENT_QUOTES);\n", 0},
		{"negative4", "<?php\nhtmlspecialchars(htmlspecialchars($s,ENT_QUOTES),ENT_NOQUOTES);\n", 0},
		{"negative5", "<?php\nhtmlspecialchars(htmlspecialchars($s),double_encode:$flag);\n", 0},
		{"negative6", "<?php\ntrim($s);\n", 0},
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
