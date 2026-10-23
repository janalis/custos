package lsp

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"sort"
	"sync"
	"testing"
	"time"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/catalogue"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/testing/testbudget"
)

// percentiles returns the median and 95th percentile of d (sorted in place).
func percentiles(d []time.Duration) (time.Duration, time.Duration) {
	sort.Slice(d, func(i, j int) bool { return d[i] < d[j] })
	return d[len(d)/2], d[len(d)*95/100]
}

// publishClock records when the first diagnostics written after a reset
// reach the client.
type publishClock struct {
	mu sync.Mutex
	at time.Time
}

func (c *publishClock) Write(p []byte) (int, error) {
	c.mu.Lock()
	if c.at.IsZero() && bytes.Contains(p, []byte("textDocument/publishDiagnostics")) {
		c.at = time.Now()
	}
	c.mu.Unlock()
	return len(p), nil
}

func (c *publishClock) reset() {
	c.mu.Lock()
	c.at = time.Time{}
	c.mu.Unlock()
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
		start, startCPU := time.Now(), testbudget.CPU()
		f := syntax.ParseBest(path, src, syntax.Options{Version: phpversion.PHP84})
		e.Analyze(f)
		wall, cpu = append(wall, time.Since(start)), append(cpu, testbudget.CPU()-startCPU)
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
// the wall-clock time until the file's diagnostics are published and until
// its flow summaries are updated too, and the CPU time of both; no budget is
// enforced.
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
	clock := &publishClock{}
	s := whiteBox(t, root, clock, catalogue.All()...)
	if s.beginIndexing() {
		s.buildIndex()
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	uri := "file://" + abs
	s.docs[uri] = &document{uri: uri, path: abs, text: src, lines: syntax.NewLineIndex(src), analyzed: -1}
	var published, wall, cpu []time.Duration
	for i := 0; i < 20; i++ {
		// A trailing comment changes the text, not the declarations.
		s.mu.Lock()
		d := s.docs[uri]
		d.text = append(append([]byte(nil), src...), []byte("\n// edit "+string(rune('a'+i%26)))...)
		d.revision++
		d.version++
		d.lines = syntax.NewLineIndex(d.text)
		s.mu.Unlock()
		clock.reset()
		start, startCPU := time.Now(), testbudget.CPU()
		s.analyzeChanged(uri)
		wall, cpu = append(wall, time.Since(start)), append(cpu, testbudget.CPU()-startCPU)
		published = append(published, clock.at.Sub(start))
	}
	d50, d95 := percentiles(published)
	p50, p95 := percentiles(wall)
	c50, c95 := percentiles(cpu)
	t.Logf("edit in %s: diagnostics p50 %v, p95 %v; with the flow update p50 %v, p95 %v, CPU p50 %v, p95 %v (flow %v)", filepath.Base(path), d50, d95, p50, p95, c50, c95, s.flowEnabled())
}

// labelled runs a rule under a pprof goroutine label naming it, so that a
// CPU profile of TestRuleCost splits the analysis by rule
// (`go tool pprof -tags`). The time a rule spends building semantic state
// it is the first to ask for (types, flow) is charged to it.
type labelled struct {
	analysis.Rule
	ctx context.Context
}

func (l labelled) Check(ctx *analysis.Context, n syntax.Node) {
	pprof.SetGoroutineLabels(l.ctx)
	l.Rule.Check(ctx, n)
	pprof.SetGoroutineLabels(context.Background())
}

// labelledFile is labelled for a rule that also checks whole files.
type labelledFile struct{ labelled }

func (l labelledFile) CheckFile(ctx *analysis.Context) {
	pprof.SetGoroutineLabels(l.ctx)
	l.Rule.(analysis.FileRule).CheckFile(ctx)
	pprof.SetGoroutineLabels(context.Background())
}

// TestRuleCost analyses CUSTOS_BENCH_FILE repeatedly with every rule
// labelled, for a CPU profile (-cpuprofile) to attribute by rule:
//
//	go test -run TestRuleCost -cpuprofile cpu.out ./internal/editor/lsp
//	go tool pprof -tags cpu.out
func TestRuleCost(t *testing.T) {
	timingMeaningful(t)
	path := os.Getenv("CUSTOS_BENCH_FILE")
	if path == "" {
		t.Skip("set CUSTOS_BENCH_FILE")
	}
	src, err := os.ReadFile(path)
	if err != nil {
		t.Skip(err)
	}
	var rules []analysis.Rule
	for _, r := range catalogue.All() {
		l := labelled{r, pprof.WithLabels(context.Background(), pprof.Labels("rule", r.ID()))}
		if _, ok := r.(analysis.FileRule); ok {
			rules = append(rules, labelledFile{l})
		} else {
			rules = append(rules, l)
		}
	}
	e, err := analysis.NewEngine(rules, analysis.Config{PHP: phpversion.PHP84})
	if err != nil {
		t.Fatal(err)
	}
	for range 40 {
		e.Analyze(syntax.ParseBest(path, src, syntax.Options{Version: phpversion.PHP84}))
	}
}

// BenchmarkEditAnalysis is TestEditLatency as a benchmark (parse and every
// default rule on CUSTOS_BENCH_FILE), for allocation counts, which do not
// depend on the machine's load.
func BenchmarkEditAnalysis(b *testing.B) {
	path := os.Getenv("CUSTOS_BENCH_FILE")
	if path == "" {
		b.Skip("set CUSTOS_BENCH_FILE")
	}
	src, err := os.ReadFile(path)
	if err != nil {
		b.Skip(err)
	}
	e, err := analysis.NewEngine(catalogue.All(), analysis.Config{PHP: phpversion.PHP84})
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	for b.Loop() {
		e.Analyze(syntax.ParseBest(path, src, syntax.Options{Version: phpversion.PHP84}))
	}
}
