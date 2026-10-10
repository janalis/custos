package processclosedbeforeownedpipes

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ProcessClosedBeforeOwnedPipes"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"proc_close(makeProcess());", 0},
		{"proc_close($p);", 0},
		{"echo 1;proc_close($p);", 0},
		{"$other=proc_open(\"x\",[], $pipes);proc_close($p);", 0},
		{"$p=makeProcess();proc_close($p);", 0},
		{"$p=proc_open(\"x\",[],$object->pipes);proc_close($p);", 0},
		{"$p=proc_open(\"x\",$descriptors,$pipes);proc_close($p);", 0},
		{"$p=proc_open(\"x\",[1=>$unknown],$pipes);proc_close($p);", 0},
		{"$p=proc_open(\"x\",[1=>[$kind,\"w\"]],$pipes);proc_close($p);", 0},
		{"$p=proc_open(\"x\",[1=>[\"pipe\",\"invalid\"]],$pipes);proc_close($p);", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
