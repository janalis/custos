package immutabledateresultignored_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/immutabledateresultignored"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$day=new DateTimeImmutable('2025-03-10'); $day->modify('+1 day'); echo $day->format('c');\n", 1},
		{"negative0", "<?php\n$x=new DateTime();$x->modify('+1 day');\n", 0},
		{"negative1", "<?php\n$x=new DateTimeImmutable();$x=$x->modify('+1 day');\n", 0},
		{"negative2", "<?php\n(new DateTimeImmutable())->format('c');\n", 0},
		{"negative3", "<?php\n$x->modify('+1 day');\n", 0},
		{"negative4", "<?php\n$x=new DateTimeImmutable();$x->{$name}();\n", 0},
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

func TestLegacyVersion(t *testing.T) {
	e, err := analysis.NewEngine([]analysis.Rule{rule.New()}, analysis.Config{PHP: phpversion.PHP53, Only: []string{rule.New().ID()}})
	if err != nil {
		t.Fatal(err)
	}
	f := e.Analyze(syntax.Parse("old.php", []byte("<?php\n$day=new DateTimeImmutable('2025-03-10'); $day->modify('+1 day'); echo $day->format('c');"), syntax.Options{}))
	if len(f) != 0 {
		t.Fatalf("legacy findings: %+v", f)
	}
}
