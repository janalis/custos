package selectwatcharraysnotrestored

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkAuditProof(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"SelectWatchArraysNotRestored"}})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte(`<?php $read=[$s];$write=null;$except=null;while($run){stream_select($read,$write,$except,1);}`)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(syntax.Parse("audit.php", src, syntax.Options{}))
	}
}
