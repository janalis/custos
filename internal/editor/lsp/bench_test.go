package lsp

import (
	"os"
	"sort"
	"testing"
	"time"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/catalogue"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// TestEditLatency measures the per-keystroke re-analysis cost (parse +
// all enabled-by-default rules) on a large real file (CUSTOS_BENCH_FILE,
// e.g. symfony/console's Application.php); budget p95 < 30ms.
// Wall-clock budgets fail spuriously on a loaded machine, so the budget is
// only enforced with CUSTOS_PERF=1 (`make bench`); otherwise it is logged.
func TestEditLatency(t *testing.T) {
	if testing.Short() || raceEnabled || testing.CoverMode() != "" {
		t.Skip("short mode, race detector or coverage (timing not meaningful)")
	}
	path := os.Getenv("CUSTOS_BENCH_FILE")
	if path == "" {
		t.Skip("set CUSTOS_BENCH_FILE")
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Skip(err)
	}
	e, err := analysis.NewEngine(catalogue.All(), analysis.Config{PHP: phpversion.PHP84})
	if err != nil {
		t.Fatal(err)
	}
	var d []time.Duration
	for i := 0; i < 40; i++ {
		start := time.Now()
		f := syntax.ParseBest(path, src, syntax.Options{Version: phpversion.PHP84})
		e.Analyze(f)
		d = append(d, time.Since(start))
	}
	sort.Slice(d, func(i, j int) bool { return d[i] < d[j] })
	p50, p95 := d[len(d)/2], d[len(d)*95/100]
	t.Logf("%d KB file: p50 %v, p95 %v (%d rules)", len(src)/1024, p50, p95, len(e.Rules()))
	if p95 > 30*time.Millisecond && os.Getenv("CUSTOS_PERF") == "1" {
		t.Errorf("p95 %v exceeds the 30ms budget", p95)
	}
}
