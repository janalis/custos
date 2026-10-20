package lsp

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"syscall"
	"testing"
	"time"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/catalogue"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
)

// cpuTime is the CPU time (user + system) the process used so far: unlike
// wall-clock time, it does not grow while other processes hold the CPU.
func cpuTime(t *testing.T) time.Duration {
	t.Helper()
	var ru syscall.Rusage
	if err := syscall.Getrusage(syscall.RUSAGE_SELF, &ru); err != nil {
		t.Fatal(err)
	}
	return time.Duration(ru.Utime.Nano() + ru.Stime.Nano())
}

// percentiles returns the median and 95th percentile of d (sorted in place).
func percentiles(d []time.Duration) (time.Duration, time.Duration) {
	sort.Slice(d, func(i, j int) bool { return d[i] < d[j] })
	return d[len(d)/2], d[len(d)*95/100]
}

// timingMeaningful skips performance tests where timings mean nothing.
func timingMeaningful(t *testing.T) {
	t.Helper()
	if testing.Short() || raceEnabled || testing.CoverMode() != "" {
		t.Skip("short mode, race detector or coverage (timing not meaningful)")
	}
}

// TestEditLatency measures the per-keystroke re-analysis cost (parse +
// all enabled-by-default rules) on a large real file (CUSTOS_BENCH_FILE,
// e.g. symfony/console's Application.php); budget p95 < 30ms.
// Wall-clock budgets fail spuriously on a loaded machine, so the budget is
// only enforced with CUSTOS_PERF=1 (`make bench`); otherwise it is logged.
// The CPU time of each analysis, which a loaded machine does not inflate,
// is logged next to it.
func TestEditLatency(t *testing.T) {
	timingMeaningful(t)
	path := os.Getenv("CUSTOS_BENCH_FILE")
	if path == "" {
		t.Skip("set CUSTOS_BENCH_FILE")
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Skip(err)
	}
	// The server samples no heap allocations (see cli.Run).
	defer func(rate int) { runtime.MemProfileRate = rate }(runtime.MemProfileRate)
	runtime.MemProfileRate = 0
	e, err := analysis.NewEngine(catalogue.All(), analysis.Config{PHP: phpversion.PHP84})
	if err != nil {
		t.Fatal(err)
	}
	var wall, cpu []time.Duration
	for i := 0; i < 40; i++ {
		start, startCPU := time.Now(), cpuTime(t)
		f := syntax.ParseBest(path, src, syntax.Options{Version: phpversion.PHP84})
		e.Analyze(f)
		wall, cpu = append(wall, time.Since(start)), append(cpu, cpuTime(t)-startCPU)
	}
	p50, p95 := percentiles(wall)
	c50, c95 := percentiles(cpu)
	t.Logf("%d KB file: p50 %v, p95 %v; CPU p50 %v, p95 %v (%d rules)", len(src)/1024, p50, p95, c50, c95, len(e.Rules()))
	if p95 > 30*time.Millisecond && os.Getenv("CUSTOS_PERF") == "1" {
		t.Errorf("p95 %v exceeds the 30ms budget", p95)
	}
}

// TestWorkspaceEditLatency measures what an edit costs the server in a
// project (CUSTOS_BENCH_ROOT, holding CUSTOS_BENCH_FILE): every rule, the
// project index and flow snapshot, and a body-only change to the open
// file analysed as the debounce timer would (without the delay). It logs
// wall-clock and CPU percentiles; no budget is enforced.
func TestWorkspaceEditLatency(t *testing.T) {
	timingMeaningful(t)
	root, path := os.Getenv("CUSTOS_BENCH_ROOT"), os.Getenv("CUSTOS_BENCH_FILE")
	if root == "" || path == "" {
		t.Skip("set CUSTOS_BENCH_ROOT and CUSTOS_BENCH_FILE")
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Skip(err)
	}
	s := whiteBox(t, root, io.Discard, catalogue.All()...)
	if s.beginIndexing() {
		s.buildIndex()
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	uri := "file://" + abs
	s.docs[uri] = &document{uri: uri, path: abs, text: src, lines: syntax.NewLineIndex(src), analyzed: -1}
	var wall, cpu []time.Duration
	for i := 0; i < 20; i++ {
		// A trailing comment changes the text, not the declarations.
		s.mu.Lock()
		d := s.docs[uri]
		d.text = append(append([]byte(nil), src...), []byte("\n// edit "+string(rune('a'+i%26)))...)
		d.revision++
		d.version++
		d.lines = syntax.NewLineIndex(d.text)
		s.mu.Unlock()
		start, startCPU := time.Now(), cpuTime(t)
		s.analyzeChanged(uri)
		wall, cpu = append(wall, time.Since(start)), append(cpu, cpuTime(t)-startCPU)
	}
	p50, p95 := percentiles(wall)
	c50, c95 := percentiles(cpu)
	t.Logf("edit in %s: p50 %v, p95 %v; CPU p50 %v, p95 %v (flow %v)", filepath.Base(path), p50, p95, c50, c95, s.flowEnabled())
}
