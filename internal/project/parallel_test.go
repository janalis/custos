package project

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"custos/internal/diagnostic"
	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
)

func TestParallelIsolatesPanics(t *testing.T) {
	var done atomic.Int32
	var mu sync.Mutex
	crashed := map[int]any{}
	err := parallel(context.Background(), 20, 3, func(i int) {
		if i%7 == 3 {
			panic(i)
		}
		done.Add(1)
	}, func(i int, r any) {
		mu.Lock()
		defer mu.Unlock()
		crashed[i] = r
	})
	if err != nil || done.Load() != 17 || len(crashed) != 3 || crashed[10] != 10 {
		t.Fatalf("err %v, done %d, crashed %v", err, done.Load(), crashed)
	}
	// Without a handler the crashed items are only left out.
	done.Store(0)
	if err := parallel(context.Background(), 4, 0, func(i int) {
		if i == 0 {
			panic("dropped")
		}
		done.Add(1)
	}, nil); err != nil || done.Load() != 3 {
		t.Fatalf("err %v, done %d", err, done.Load())
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := parallel(ctx, 4, 0, func(int) { t.Error("started after cancellation") }, nil); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
}

func TestEachReportsTheFirstCrash(t *testing.T) {
	var done atomic.Int32
	err := Each(context.Background(), 8, func(i int) {
		if i == 5 {
			panic("boom")
		}
		done.Add(1)
	})
	if err == nil || !strings.Contains(err.Error(), "internal failure on item 5: boom") || done.Load() != 7 {
		t.Fatalf("err %v, done %d", err, done.Load())
	}
	if err := Each(context.Background(), 2, func(int) {}); err != nil {
		t.Fatal(err)
	}
}

type crashingAnalyzer struct{}

func (crashingAnalyzer) Analyze(*syntax.File) []diagnostic.Finding { panic("engine bug") }

func TestAnalyzeBufferReportsACrashAsIncomplete(t *testing.T) {
	src := []byte("<?php echo 1;")
	result := AnalyzeBuffer(crashingAnalyzer{}, "a.php", src, syntax.Options{})
	if result.Path != "a.php" || string(result.Src) != string(src) || len(result.Findings) != 1 || result.Findings[0].Rule != "internal" ||
		!strings.Contains(result.Findings[0].Message, "engine bug") || len(result.Incomplete) != 1 || result.Incomplete[0].Reason != "internal-failure" {
		t.Fatalf("%+v", result)
	}
}

func TestFlowSurvivesSummaryPanic(t *testing.T) {
	root := t.TempDir()
	good := filepath.Join(root, "good.php")
	bad := filepath.Join(root, "bad.php")
	writeFile(t, good, "<?php function good() { return $_GET['v']; }")
	writeFile(t, bad, "<?php function bad() { return $_GET['v']; }")
	opt := syntax.Options{Version: phpversion.Default}
	ix := BuildIndex([]string{good, bad}, opt, infer.Options{})
	old := extractFlow
	t.Cleanup(func() { extractFlow = old })
	extractFlow = func(f *syntax.File, ix *index.Index, v phpversion.Version, s *flow.Snapshot, o flow.Options) *flow.File {
		if strings.HasSuffix(f.Path, "bad.php") {
			panic("simulated summary crash")
		}
		return old(f, ix, v, s, o)
	}
	build := BuildFlow(context.Background(), []string{good, bad}, nil, opt, ix, flow.Options{})
	if build.results[0] == nil || build.results[1] != nil {
		t.Fatalf("summaries: %v", build.results)
	}
	next, ok := build.Update(map[string][]byte{bad: []byte("<?php function bad() { return 1; }")}, ix)
	if !ok || next.results[0] == nil || next.results[1] != nil {
		t.Fatal("an edit of the crashing file must keep the others' summaries")
	}
}

func TestPrepareFixesSurvivesFixCrash(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a.php")
	writeFile(t, path, "<?php echo 1;")
	old := fixSource
	t.Cleanup(func() { fixSource = old })
	fixSource = func(fixing.Analyzer, string, []byte, fixing.Options) fixing.Result { panic("fix bug") }
	p, err := Open([]analysis.Rule{untrustedshellcommand.New()}, Options{Root: root, Paths: []string{root}, Analysis: analysis.Config{EnableAll: true, PHP: phpversion.Default}, Parse: syntax.Options{Version: phpversion.Default}})
	if err != nil {
		t.Fatal(err)
	}
	outs := p.PrepareFixes()
	if len(outs) != 1 || outs[0].Err == nil || !strings.Contains(outs[0].Err.Error(), "internal failure, left unchanged: fix bug") || outs[0].Output != nil {
		t.Fatalf("%+v", outs)
	}
}

func TestPrepareFixesReusesReadSources(t *testing.T) {
	root := t.TempDir()
	path, link := filepath.Join(root, "a.php"), filepath.Join(root, "link.php")
	writeFile(t, path, "<?php echo 'disk';")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	old := fixSource
	t.Cleanup(func() { fixSource = old })
	var seen []string
	fixSource = func(_ fixing.Analyzer, _ string, src []byte, _ fixing.Options) fixing.Result {
		seen = append(seen, string(src))
		return fixing.Result{Source: src}
	}
	engine, err := analysis.NewEngine(nil, analysis.Config{})
	if err != nil {
		t.Fatal(err)
	}
	outs := PrepareFixes(engine, []string{path, link}, [][]byte{[]byte("<?php echo 'kept';"), []byte("<?php")}, syntax.Options{})
	if len(seen) != 1 || seen[0] != "<?php echo 'kept';" || !errors.Is(outs[1].Err, ErrSkipped) {
		t.Fatalf("seen=%q outs=%+v", seen, outs)
	}
}
