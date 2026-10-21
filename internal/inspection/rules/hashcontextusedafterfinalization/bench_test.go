package hashcontextusedafterfinalization

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkRepeatedUpdates(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"HashContextUsedAfterFinalization"}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("bench.php", []byte("<?php $h=hash_init('sha256');hash_final($h);"+strings.Repeat("hash_update($h,'x');", 100)), syntax.Options{})
	e.Analyze(file)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}
