package isoweekcalendaryearmismatch_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/isoweekcalendaryearmismatch"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$day=new DateTimeImmutable(); $label=$day->format('Y-W');\n", 1},
		{"negative0", "<?php\ndate('o-W');\n", 0},
		{"negative1", "<?php\ndate('Y-m-d W');\n", 0},
		{"negative2", "<?php\ndate('Y');\n", 0},
		{"negative3", "<?php\ndate('\\\\Y-W');\n", 0},
		{"negative4", "<?php\n$x->format('Y-W');\n", 0},
		{"extra0", "<?php\ndate('Y-W \\\\x');\n", 1},
		{"extra1", "<?php\n$fmt='Y-W';date($fmt);\n", 1},
		{"escaped retained token", "<?php date('Y-W \\x');", 1},
		{"month labelled by year", "<?php date('F Y, \\w\\e\\e\\k W'); date('Y-m W'); date('n Y W'); date('M Y W');", 0},
		{"day labelled by year", "<?php date('Y/z W'); date('Y W d'); date('Y W j');", 0},
		{"year facts", "<?php date('L Y W'); date('t Y W');", 0},
		{"composite date", "<?php gmdate('c W Y'); gmdate('r W Y'); gmdate('Y W U');", 0},
		{"escaped calendar tokens", "<?php date('\\m\\d\\c Y-W');", 1},
		{"weekday and timezone", "<?php date('Y W N'); date('W Y D'); date('Y-W T e');", 3},
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
