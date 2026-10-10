package semanticquery

import (
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

func BenchmarkExpansionLanguageLifecycle(b *testing.B) {
	p := expansionCryptoDBProbe{id: "FiberStartedTwice", kinds: []syntax.NodeKind{syntax.KMethodCall}, check: func(ctx *analysis.Context, n syntax.Node) { CheckFiberStartedTwice(ctx, n, "finding") }}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
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

func BenchmarkExpansionLanguageFallthrough(b *testing.B) {
	p := expansionCryptoDBProbe{id: "NonVoidFunctionFallsThrough", kinds: []syntax.NodeKind{syntax.KFunction}, check: func(ctx *analysis.Context, n syntax.Node) { CheckNonVoidFunctionFallsThrough(ctx, n, "finding") }}
	e, err := analysis.NewEngine([]analysis.Rule{p}, analysis.Config{Only: []string{p.id}})
	if err != nil {
		b.Fatal(err)
	}
	file := syntax.Parse("bench.php", []byte("<?php function f(): int { if ($unknown) { 1 / 0; } else { echo 'x'; } }"), syntax.Options{})
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(file)
	}
}
