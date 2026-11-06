package lsp

import (
	"os"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/catalogue"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// BenchmarkEditAllocations complements the latency gate with warmed per-edit
// allocation counts; initialization of the embedded stubs is outside the loop.
func BenchmarkEditAllocations(b *testing.B) {
	path := os.Getenv("CUSTOS_BENCH_FILE")
	if path == "" {
		b.Skip("set CUSTOS_BENCH_FILE")
	}
	src, err := os.ReadFile(path)
	if err != nil {
		b.Fatal(err)
	}
	engine, err := analysis.NewEngine(catalogue.All(), analysis.Config{PHP: phpversion.PHP84})
	if err != nil {
		b.Fatal(err)
	}
	engine.Analyze(syntax.ParseBest(path, src, syntax.Options{Version: phpversion.PHP84}))
	b.ReportAllocs()
	for b.Loop() {
		engine.Analyze(syntax.ParseBest(path, src, syntax.Options{Version: phpversion.PHP84}))
	}
}
