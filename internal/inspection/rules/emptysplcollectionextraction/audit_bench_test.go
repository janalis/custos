package emptysplcollectionextraction

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkAuditProof(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"EmptySplCollectionExtraction"}})
	if err != nil {
		b.Fatal(err)
	}
	src := []byte(`<?php $q=new SplQueue();$q->dequeue();`)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(syntax.Parse("audit.php", src, syntax.Options{}))
	}
}
