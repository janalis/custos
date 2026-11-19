package fiberstartedtwice

import (
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
)

func BenchmarkLifecycle(b *testing.B) {
	e, err := analysis.NewEngine([]analysis.Rule{New()}, analysis.Config{Only: []string{"FiberStartedTwice"}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("bench.php", []byte("<?php $f=new Fiber(fn()=>Fiber::suspend());"+strings.Repeat("$f->start();", 100)), syntax.Options{})
	e.Analyze(file)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		e.Analyze(file)
	}
}
