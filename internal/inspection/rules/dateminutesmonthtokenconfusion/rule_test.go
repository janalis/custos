package dateminutesmonthtokenconfusion_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/dateminutesmonthtokenconfusion"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$day=new DateTimeImmutable(); $clock=$day->format('H:m:s');\n", 1},
		{"negative0", "<?php\ndate('H:i:s');\n", 0},
		{"negative1", "<?php\ndate('Y-m-d H:m:s');\n", 0},
		{"negative2", "<?php\ndate('\\\\H:m:s');\n", 0},
		{"negative3", "<?php\ndate('m');\n", 0},
		{"negative4", "<?php\ndate($format);\n", 0},
		{"extra0", "<?php\ndate('H:m:s \\\\x');\n", 1},
		{"extra1", "<?php\n$fmt='H:m:s';date($fmt);\n", 1},
		{"escaped retained token", "<?php date('H:m:s \\x foo');", 1},
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
