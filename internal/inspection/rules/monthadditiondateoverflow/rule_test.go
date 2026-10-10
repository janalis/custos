package monthadditiondateoverflow_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/monthadditiondateoverflow"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$next=(new DateTimeImmutable('2025-01-31'))->modify('+1 month');\n", 1},
		{"negative0", "<?php\n(new DateTimeImmutable('2025-01-28'))->modify('+1 month');\n", 0},
		{"negative1", "<?php\n(new DateTimeImmutable($date))->modify('+1 month');\n", 0},
		{"negative2", "<?php\n(new DateTimeImmutable('2025-01-31'))->modify('+1 day');\n", 0},
		{"negative3", "<?php\n(new DateTimeImmutable('invalid'))->modify('+1 month');\n", 0},
		{"negative4", "<?php\n$x->modify('+1 month');\n", 0},
		{"negative5", "<?php\n$x=DateTimeImmutable::createFromFormat('!Y-m-d',$date);$x->modify('+1 month');\n", 0},
		{"negative6", "<?php\n(new DateTimeImmutable('2025-03-28'))->modify('-1 month');\n", 0},
		{"mutable date changed", "<?php $d=new DateTime('2025-01-31');$d->modify('first day of this month');$d->modify('+1 month');", 0},
		{"fresh mutable date", "<?php (new DateTime('2025-01-31'))->modify('+1 month');", 1},
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
