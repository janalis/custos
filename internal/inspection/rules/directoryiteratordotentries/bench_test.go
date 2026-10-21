package directoryiteratordotentries

import (
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkIteratorBindingProof(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"DirectoryIteratorDotEntries"}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("iterator.php", []byte("<?php foreach(new DirectoryIterator('/tmp') as $entry){if($entry->isDot()){continue;}unlink($entry->getPathname());}"), syntax.Options{})
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		e.Analyze(file)
	}
}
