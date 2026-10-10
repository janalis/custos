package processpipesdrainedsequentially

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func TestOperationalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ProcessPipesDrainedSequentially"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{
		{"$p=proc_open(\"x\",[1=>[\"pipe\",\"w\"],2=>[\"pipe\",\"w\"]],$pipes);echo 1;stream_get_contents($pipes[2]);", 0},
		{"$p=proc_open(\"x\",[1=>[\"pipe\",\"w\"],2=>[\"pipe\",\"w\"]],$pipes);$x=1;stream_get_contents($pipes[2]);", 0},
		{"$p=proc_open(\"x\",[1=>[\"pipe\",\"w\"],2=>[\"pipe\",\"w\"]],$pipes);stream_get_contents($other);stream_get_contents($pipes[2]);", 0},
		{"$p=proc_open(\"x\",[1=>[\"pipe\",\"w\"],2=>[\"pipe\",\"w\"]],$pipes);stream_get_contents($other[1]);stream_get_contents($pipes[2]);", 0},
		{"$p=other();stream_get_contents($pipes[1]);stream_get_contents($pipes[2]);", 0},
		{"$p=proc_open(\"x\",[1=>[\"pipe\",\"w\"],2=>[\"pipe\",\"w\"]],$other);stream_get_contents($pipes[1]);stream_get_contents($pipes[2]);", 0},
		{"$p=proc_open(\"x\",$unknown,$pipes);stream_get_contents($pipes[1]);stream_get_contents($pipes[2]);", 0},
		{"$p=proc_open(\"x\",[1=>[\"pipe\",\"w\"]],$pipes);stream_get_contents($pipes[1]);stream_get_contents($pipes[2]);", 0},
		{"$p=proc_open(\"x\",[1=>[\"file\",\"/tmp/x\",\"w\"],2=>[\"pipe\",\"w\"]],$pipes);stream_get_contents($pipes[1]);stream_get_contents($pipes[2]);", 0},
		{"echo 1;stream_get_contents($pipes[1]);stream_get_contents($pipes[2]);", 0},
	} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}

func TestOperationalAdditionalBoundaries(t *testing.T) {
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"ProcessPipesDrainedSequentially"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		source string
		want   int
	}{{"stream_get_contents($pipes[2]);", 0}} {
		t.Run(tc.source, func(t *testing.T) {
			got := engine.Analyze(syntax.Parse("case.php", []byte("<?php "+tc.source), syntax.Options{}))
			if len(got) != tc.want {
				t.Fatalf("got %d want %d: %+v", len(got), tc.want, got)
			}
		})
	}
}
