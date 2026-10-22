package lsp

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/unescapedhtmloutput"
	"custos/internal/php/syntax"
	"custos/internal/project/config"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/testing/testbudget"
)

func TestAssertionOverrideFalse(t *testing.T) {
	enabled := true
	base, err := config.Resolve(t.TempDir(), config.File{Flow: &config.FlowSettings{Assertions: &enabled}})
	if err != nil {
		t.Fatal(err)
	}
	for _, over := range []config.File{{}, {Flow: &config.FlowSettings{}}, {Flow: &config.FlowSettings{Assertions: new(bool)}}} {
		merged, err := mergeOverrides(base, over)
		want := over.Flow == nil || over.Flow.Assertions == nil
		if err != nil || merged.Flow.Assertions != want {
			t.Fatalf("override=%+v merged=%+v err=%v", over, merged, err)
		}
	}
	empty := config.File{Flow: &config.FlowSettings{}}
	first, err := mergeOverrides(base, empty)
	if err != nil || !first.Flow.Assertions {
		t.Fatal("first inherited mode")
	}
	base.Flow.Assertions = false
	second, err := mergeOverrides(base, empty)
	if err != nil || second.Flow.Assertions || empty.Flow.Assertions != nil {
		t.Fatal("empty override retained or mutated prior base")
	}
}

func TestAssertionModeOwnsIndexBuild(t *testing.T) {
	for _, failOldDiscovery := range []bool{false, true} {
		t.Run(map[bool]string{false: "stale completed build", true: "stale failed discovery"}[failOldDiscovery], func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, "source.php")
			if err := os.WriteFile(path, []byte("<?php class Original {}"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(root, config.FileName), []byte(`{"flow":{"assertions":true}}`), 0o644); err != nil {
				t.Fatal(err)
			}
			s := whiteBox(t, root, io.Discard, unescapedhtmloutput.New())
			oldDiscover, oldBuild := discoverIndexFiles, buildIndexSymbols
			defer func() { discoverIndexFiles = oldDiscover; buildIndexSymbols = oldBuild }()
			oldStarted, releaseOld := make(chan struct{}), make(chan struct{})
			latestStarted, releaseLatest := make(chan struct{}, 2), make(chan struct{})
			var discoveries atomic.Int32
			discoverIndexFiles = func(paths, exclude []string) ([]string, error) {
				if discoveries.Add(1) == 1 && failOldDiscovery {
					close(oldStarted)
					<-releaseOld
					return nil, errors.New("old discovery")
				}
				return oldDiscover(paths, exclude)
			}
			buildIndexSymbols = func(root string, files []string, opt syntax.Options, options infer.Options) *index.Index {
				if options.Assertions {
					close(oldStarted)
					<-releaseOld
				} else {
					latestStarted <- struct{}{}
					<-releaseLatest
				}
				return oldBuild(root, files, opt, options)
			}
			wait := func(ch <-chan struct{}) {
				t.Helper()
				select {
				case <-ch:
				case <-time.After(testbudget.Of(5 * time.Second)):
					t.Fatal("index build did not reach checkpoint")
				}
			}
			if !s.beginIndexing() {
				t.Fatal("index not requested")
			}
			oldDone := make(chan struct{})
			go func() { s.buildIndex(); close(oldDone) }()
			wait(oldStarted)
			s.initOpts = []byte(`{"flow":{"assertions":false}}`)
			if err := s.configure(); err != nil {
				t.Fatal(err)
			}
			wait(latestStarted)
			// A duplicate configuration must restart ownership rather than strand it.
			if err := s.configure(); err != nil {
				t.Fatal(err)
			}
			wait(latestStarted)
			if err := os.WriteFile(path, []byte("<?php class Changed {}"), 0o644); err != nil {
				t.Fatal(err)
			}
			s.watchedFilesChanged([]fileChange{{URI: "file://" + path, Type: 2}})
			close(releaseOld)
			wait(oldDone)
			s.mu.Lock()
			indexing, queued := s.indexing, len(s.pending)
			s.mu.Unlock()
			if !indexing || queued != 1 {
				t.Fatalf("stale build stole latest state: indexing=%v queued=%d", indexing, queued)
			}
			close(releaseLatest)
			deadline := time.Now().Add(testbudget.Of(5 * time.Second))
			for time.Now().Before(deadline) {
				s.mu.Lock()
				done := !s.indexing && s.index != nil && s.index.Class("Changed", s.cfg.PHP) != nil && len(s.pending) == 0 && !s.cfg.Flow.Assertions
				s.mu.Unlock()
				if done {
					return
				}
				time.Sleep(time.Millisecond)
			}
			t.Fatal("latest build did not replay queued file change")
		})
	}
}

func TestAssertionModeRefreshesDependentUnsavedBuffers(t *testing.T) {
	for _, reverse := range []bool{false, true} {
		root := t.TempDir()
		s := whiteBox(t, root, io.Discard, unescapedhtmloutput.New())
		caller, callee := "a.php", "z.php"
		if reverse {
			caller, callee = callee, caller
		}
		for name, text := range map[string]string{caller: `<?php function outer($v){return inner($v);}`, callee: `<?php function inner($v){assert(($v=htmlspecialchars($v))!=='');return $v;}`} {
			path := filepath.Join(root, name)
			s.docs["file://"+path] = &document{path: path, text: []byte(text)}
		}
		for _, enabled := range []bool{false, true, false} {
			s.cfg.Flow.Assertions = enabled
			e, err := analysis.NewEngine([]analysis.Rule{unescapedhtmloutput.New()}, s.cfg.Analysis())
			if err != nil {
				t.Fatal(err)
			}
			s.engine = e
			s.index = nil
			if !s.beginIndexing() {
				t.Fatal("index not requested")
			}
			s.buildIndex()
			f := syntax.Parse("main.php", []byte(`<?php echo '<p>'.outer($_GET['x']);`), syntax.Options{Version: s.cfg.PHP})
			want := 1
			if enabled {
				want = 0
			}
			if got := len(s.engine.Analyze(f)); got != want {
				t.Fatalf("reverse=%v enabled=%v findings=%d want=%d", reverse, enabled, got, want)
			}
		}
	}
}

func TestAssertionReconfigureCancelsPendingWithoutIndex(t *testing.T) {
	s := whiteBox(t, t.TempDir(), io.Discard, unescapedhtmloutput.New())
	s.indexing = true
	s.pending = []fileChange{{URI: "file:///deleted.php", Type: 3}}
	useRules(t)
	if err := s.configure(); err != nil {
		t.Fatal(err)
	}
	if s.indexing || len(s.pending) != 0 {
		t.Fatal("cancelled indexing retained obsolete watch events")
	}
}

func TestAssertionSameModeConfigurationRetainsProjectFlow(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "source.php")
	if err := os.WriteFile(path, []byte(`<?php function source(){return $_GET['x'];}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s := whiteBox(t, root, io.Discard, unescapedhtmloutput.New())
	if !s.beginIndexing() {
		t.Fatal("index not requested")
	}
	s.buildIndex()
	if err := s.configure(); err != nil {
		t.Fatal(err)
	}
	f := syntax.Parse("caller.php", []byte(`<?php echo '<p>'.source();`), syntax.Options{})
	if got := s.engine.Analyze(f); len(got) != 1 {
		t.Fatalf("same-mode configuration lost project flow: %v", got)
	}
}

func TestIndexBuildCrashLeavesTheServerServing(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.php"), []byte("<?php class Original {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := whiteBox(t, root, io.Discard, unescapedhtmloutput.New())
	oldBuild := buildIndexSymbols
	defer func() { buildIndexSymbols = oldBuild }()
	buildIndexSymbols = func(string, []string, syntax.Options, infer.Options) *index.Index { panic("index bug") }
	if !s.beginIndexing() {
		t.Fatal("index not requested")
	}
	s.buildIndex()
	s.mu.Lock()
	indexing, ix := s.indexing, s.index
	s.mu.Unlock()
	if indexing || ix != nil {
		t.Fatal("a crashed index build must end without publishing an index")
	}
	if !s.beginIndexing() {
		t.Fatal("a later index build must be possible")
	}
}
