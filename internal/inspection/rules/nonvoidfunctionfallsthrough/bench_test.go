package nonvoidfunctionfallsthrough

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkFallthrough(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"NonVoidFunctionFallsThrough"}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("bench.php", []byte("<?php function f(): int { if ($unknown) { 1 / 0; } else { echo 'x'; } }"), syntax.Options{})
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(file)
	}
}
