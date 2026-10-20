package flow

import (
	"os"
	"testing"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
	"custos/internal/semantic/stubs"
)

// BenchmarkSummarizeFile summarizes a real file (CUSTOS_BENCH_FILE) and
// asks the value of every variable read, as the rules of an analysis do.
func BenchmarkSummarizeFile(b *testing.B) {
	path := os.Getenv("CUSTOS_BENCH_FILE")
	if path == "" {
		b.Skip("set CUSTOS_BENCH_FILE")
	}
	src, err := os.ReadFile(path)
	if err != nil {
		b.Skip(err)
	}
	f := syntax.ParseBest(path, src, syntax.Options{Version: phpversion.PHP84})
	ix := index.New(stubs.Index())
	ix.Add(index.Extract(f))
	var reads []*syntax.Variable
	syntax.InspectFile(f, func(n syntax.Node) bool {
		if v, ok := n.(*syntax.Variable); ok && v.Name != "" && v.Name != "this" {
			reads = append(reads, v)
		}
		return true
	})
	b.ReportAllocs()
	for b.Loop() {
		e, _ := Summarize(f, infer.NewEnv(f, names.New(f), ix, phpversion.PHP84), nil)
		for _, v := range reads {
			e.Value(v)
		}
	}
}
