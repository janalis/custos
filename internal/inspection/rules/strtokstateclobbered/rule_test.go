package strtokstateclobbered_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/strtokstateclobbered"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$first=strtok($a,','); strtok($b,':'); $next=strtok(',');\n", 1},
		{"negative0", "<?php\nstrtok($a,',');strtok(',');\n", 0},
		{"negative1", "<?php\nstrtok($a,',');strtok($b,':');\n", 0},
		{"negative2", "<?php\nstrtok($a,',');unknown();strtok($b,':');strtok(',');\n", 0},
		{"negative3", "<?php\nstrtok($a,',');strtok(',');strtok($b,':');strtok(',');\n", 0},
		{"same sequence restart", "<?php strtok($a,',');strtok($a,',');strtok(',');", 0},
		{"control flow barrier", "<?php strtok($a,',');if($q){strtok($b,':');}strtok(',');", 0},
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
