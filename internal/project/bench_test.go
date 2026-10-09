package project

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/catalogue"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

func BenchmarkAnalyzeBuffer(b *testing.B) {
	engine, err := analysis.NewEngine(catalogue.All(), analysis.Config{PHP: phpversion.PHP85})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte("<?php\n" + strings.Repeat("function inspect(string $s): int { return strlen($s) + 1; }\n", 30))
	opt := syntax.Options{Version: phpversion.PHP85}
	AnalyzeBuffer(engine, "buffer.php", src, opt) // warm embedded symbols
	b.ReportAllocs()
	for b.Loop() {
		AnalyzeBuffer(engine, "buffer.php", src, opt)
	}
}
