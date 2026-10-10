package intervaldaycomponentastotal_test

import (
	"testing"

	"custos/internal/inspection/analysis"
	rule "custos/internal/inspection/rules/intervaldaycomponentastotal"
	"custos/internal/php/syntax"
)

func TestContracts(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		count        int
	}{
		{"positive", "<?php\n$start=new DateTimeImmutable(); $end=new DateTimeImmutable(); $days=$start->diff($end)->format('%d');\n", 1},
		{"negative0", "<?php\n(new DateInterval('P1D'))->format('%d');\n", 0},
		{"negative1", "<?php\n$a=new DateTimeImmutable();$b=new DateTimeImmutable();$a->diff($b)->format('%a');\n", 0},
		{"negative2", "<?php\n$x->format('%d');\n", 0},
		{"negative3", "<?php\n$a=new DateTimeImmutable();$a->format('%d');\n", 0},
		{"negative4", "<?php\nclass P{function make(){return new DateInterval('P1D');}}(new P())->make()->format('%d');\n", 0},
		{"extra0", "<?php\n$fmt='%d';$a=new DateTimeImmutable();$b=new DateTimeImmutable();$a->diff($b)->format($fmt);\n", 1},
		{"mutated diff interval", "<?php $a=new DateTimeImmutable;$b=new DateTimeImmutable;$i=$a->diff($b);$i->d=0;echo $i->format('%d');", 0},
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
