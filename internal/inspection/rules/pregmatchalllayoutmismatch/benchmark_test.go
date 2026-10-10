package pregmatchalllayoutmismatch

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// BenchmarkRepeatedCaptureReads measures cached literal regex proof across many reads.
func BenchmarkRepeatedCaptureReads(b *testing.B) {
	source := []byte("<?php preg_match_all(\"/(a)/\",\"aaa\",$m);echo $m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0],$m[0][0];")
	file := syntax.Parse("bench.php", source, syntax.Options{})
	engine, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{New().ID()}})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetBytes(int64(len(source)))
	b.ResetTimer()
	for b.Loop() {
		engine.Analyze(file)
	}
}
