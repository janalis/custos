package ambiguousreplacementbackreference_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/ambiguousreplacementbackreference"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$result=preg_replace('/(r)/','$11','r');\n", 1},
		{"negative0", "<?php\npreg_replace('/(r)/','${1}1','r');\n", 0},
		{"negative1", "<?php\npreg_replace('/x/','$11','x');\n", 0},
		{"negative2", "<?php\npreg_replace($pattern,'$11','x');\n", 0},
		{"negative3", "<?php\npreg_replace('/(r)/','$1','r');\n", 0},
		{"negative4", "<?php\npreg_replace('/(?|x)/','$11','x');\n", 0},
		{"negative5", "<?php\npreg_replace('/(r)/',$replacement,'r');\n", 0},
		{"negative6", "<?php\ntrim($s);\n", 0},
		{"negative7", "<?php\npreg_replace('/(r)/','\\\\$11','r');\n", 0},
		{"extra0", "<?php\npreg_replace('/(r)/',\"\\$11\",'r');\n", 1},
		{"PCRE quoted literal span", "<?php preg_replace('/\\\\Q(a)\\\\E(b)/','$21','(a)b');", 0},
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
