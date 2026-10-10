package millisecondsusedasunixseconds_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/millisecondsusedasunixseconds"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\necho date('c',1730000000000);\n", 1},
		{"negative0", "<?php\ndate('c',1700000000);\n", 0},
		{"negative1", "<?php\ndate('c',$ts/1000);\n", 0},
		{"negative2", "<?php\ndate('c',$ts);\n", 0},
		{"negative3", "<?php\n$d->setTimestamp(1730000000000);\n", 0},
		{"negative4", "<?php\n(new DateTimeImmutable())->format('c');\n", 0},
		{"extra0", "<?php\n(new DateTimeImmutable())->setTimestamp(1730000000000);\n", 1},
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
