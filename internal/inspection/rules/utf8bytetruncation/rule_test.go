package utf8bytetruncation_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/utf8bytetruncation"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$initial=substr('ñame',0,1);\n", 1},
		{"negative0", "<?php\nsubstr('ñame',0,2);\n", 0},
		{"negative1", "<?php\nsubstr('name',0,1);\n", 0},
		{"negative2", "<?php\nsubstr($s,0,1);\n", 0},
		{"negative3", "<?php\nsubstr('ñame',$start,1);\n", 0},
		{"negative4", "<?php\nsubstr('ñame',20,1);\n", 0},
		{"negative5", "<?php\nsubstr('ñame',-100,2);\n", 0},
		{"negative6", "<?php\nsubstr('ñame',0,-3);\n", 0},
		{"negative7", "<?php\nsubstr('ñame',0,$length);\n", 0},
		{"negative8", "<?php\nsubstr('ñame',0,-100);\n", 0},
		{"negative9", "<?php\nstrlen('ñ');\n", 0},
		{"negative10", "<?php\nsubstr('ñame',5,-100);\n", 0},
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
