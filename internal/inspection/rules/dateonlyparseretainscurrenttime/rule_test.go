package dateonlyparseretainscurrenttime_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/dateonlyparseretainscurrenttime"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$day=DateTimeImmutable::createFromFormat('Y-m-d',$input);\n", 1},
		{"negative0", "<?php\nDateTimeImmutable::createFromFormat('!Y-m-d',$d);\n", 0},
		{"negative1", "<?php\nDateTimeImmutable::createFromFormat('Y-m-d|',$d);\n", 0},
		{"negative2", "<?php\nDateTimeImmutable::createFromFormat('Y-m-d H:i:s',$d);\n", 0},
		{"negative3", "<?php\nDateTimeImmutable::createFromFormat('H:i:s',$d);\n", 0},
		{"negative4", "<?php\nDateTimeImmutable::createFromFormat($format,$d);\n", 0},
		{"negative5", "<?php\nstrlen('x');\n", 0},
		{"negative6", "<?php\nUnknown::createFromFormat('Y-m-d',$date);\n", 0},
		{"extra0", "<?php\n$fmt='Y-m-d';DateTimeImmutable::createFromFormat($fmt,$date);\n", 1},
		{"escaped calendar tokens", "<?php DateTimeImmutable::createFromFormat('\\\\Y-\\\\m-\\\\d','Y-m-d');", 0},
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
