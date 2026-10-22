package project

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
)

func TestProjectComposesFlowWithoutWriting(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "wrapper.php"), "<?php function invoke($arg) { system($arg); }")
	path := filepath.Join(root, "main.php")
	src := "<?php invoke($_GET['command']);"
	writeFile(t, path, src)
	p, err := Open([]analysis.Rule{untrustedshellcommand.New()}, Options{Root: root, Paths: []string{root}, Analysis: analysis.Config{EnableAll: true, PHP: phpversion.Default}, Parse: syntax.Options{Version: phpversion.Default}})
	if err != nil {
		t.Fatal(err)
	}
	count := 0
	for _, result := range p.AnalyzeReport() {
		count += len(result.Findings)
	}
	if count != 1 {
		t.Fatalf("cross-file wrapper findings: %d", count)
	}
	for _, prepared := range p.PrepareFixes() {
		if prepared.Err != nil || prepared.Applied != 0 || prepared.Output != nil {
			t.Fatalf("unsafe automatic shell repair: %+v", prepared)
		}
	}
}

func TestFlowSnapshotSourcesAndBoundedComposition(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "wrapper.php")
	writeFile(t, path, "<?php function source() { return $_GET['value']; }")
	opt := syntax.Options{Version: phpversion.Default}
	ix := BuildIndex([]string{path}, opt, infer.Options{})
	fromDisk := BuildFlowSnapshot(context.Background(), []string{path, filepath.Join(root, "gone.php")}, nil, opt, ix, flow.Options{})
	fromBuffer := BuildFlowSnapshot(context.Background(), []string{path}, [][]byte{[]byte("<?php function source() { return $_GET['value']; }")}, opt, ix, flow.Options{})
	if !fromDisk.Equal(fromBuffer) {
		t.Fatal("disk and supplied source summaries differ")
	}
	// The functions of one file are summarized callee first: a chain of
	// them composes in one round, however long.
	var chain strings.Builder
	chain.WriteString("<?php function f0() { return $_GET['value']; }")
	for i := 1; i < 12; i++ {
		fmt.Fprintf(&chain, "function f%d() { return f%d(); }", i, i-1)
	}
	ix.Add(ExtractSymbols(path, []byte(chain.String()), opt, infer.Options{}))
	long := BuildFlowSnapshot(context.Background(), []string{path}, [][]byte{[]byte(chain.String())}, opt, ix, flow.Options{})
	if long == nil || long.Equal(fromDisk) {
		t.Fatal("composition did not retain distinct wrapper summaries")
	}
	// A chain across files takes a round per file: one longer than the
	// composition budget must stop conservatively, rather than recurse
	// indefinitely or retain parser trees in the snapshot.
	var paths []string
	var srcs [][]byte
	for i := range 12 {
		src := fmt.Sprintf("<?php function g%d() { return g%d(); }", i, i-1)
		if i == 0 {
			src = "<?php function g0() { return $_GET['value']; }"
		}
		paths = append(paths, filepath.Join(root, fmt.Sprintf("g%d.php", i)))
		srcs = append(srcs, []byte(src))
		ix.Add(ExtractSymbols(paths[i], srcs[i], opt, infer.Options{}))
	}
	if BuildFlowSnapshot(context.Background(), paths, srcs, opt, ix, flow.Options{}) == nil {
		t.Fatal("bounded composition returned no snapshot")
	}
}

// An incremental update gives the snapshot a full build gives for the same
// sources, through edits that add and remove sinks, chains of wrappers
// across files and recursion between files.
func TestFlowBuildUpdateMatchesFullBuild(t *testing.T) {
	root := t.TempDir()
	opt := syntax.Options{Version: phpversion.Default}
	sources := map[string]string{
		"wrapper.php": "<?php function invoke($a): void { system($a); }",
		"relay.php":   "<?php function relay($x): void { invoke($x); }",
		"ping.php":    "<?php function ping($n): void { if ($n) { pong($n); } exec($n); }",
		"pong.php":    "<?php function pong($n): void { ping($n); }",
		"user.php":    "<?php relay($_GET['a']); ping($_POST['b']);",
	}
	var files []string
	for name, src := range sources {
		path := filepath.Join(root, name)
		writeFile(t, path, src)
		files = append(files, path)
	}
	sort.Strings(files)
	ix := BuildIndex(files, opt, infer.Options{})
	build := BuildFlow(context.Background(), files, nil, opt, ix, flow.Options{})
	current := map[string]string{}
	for name, src := range sources {
		current[filepath.Join(root, name)] = src
	}
	edits := []struct{ name, src string }{
		{"wrapper.php", "<?php function invoke($a): void { echo $a; }"},
		{"relay.php", "<?php function relay($x): void { invoke(strtolower($x)); }"},
		{"wrapper.php", "<?php function invoke($a): void { system($a); }"},
		{"pong.php", "<?php function pong($n): void { passthru($n); }"},
		{"pong.php", "<?php\n\n// moved down\nfunction pong($n): void { ping($n); }"},
		{"relay.php", "<?php function relay($x): void { invoke($x); }"},
	}
	for step, edit := range edits {
		path := filepath.Join(root, edit.name)
		current[path] = edit.src
		ix.Add(ExtractSymbols(path, []byte(edit.src), opt, infer.Options{}))
		next, ok := build.Update(map[string][]byte{path: []byte(edit.src)}, ix)
		if !ok {
			t.Fatalf("step %d: update refused", step)
		}
		srcs := make([][]byte, len(files))
		for i, f := range files {
			srcs[i] = []byte(current[f])
		}
		if full := BuildFlow(context.Background(), files, srcs, opt, ix, flow.Options{}); !next.Snapshot.Equal(full.Snapshot) {
			t.Fatalf("step %d (%s): incremental snapshot differs from a full build", step, edit.name)
		}
		if step == 0 && next.Snapshot.Equal(build.Snapshot) {
			t.Fatal("removing a sink left the snapshot unchanged")
		}
		build = next
	}
	if _, ok := build.Update(map[string][]byte{filepath.Join(root, "new.php"): []byte("<?php")}, ix); ok {
		t.Fatal("update accepted a file outside the build")
	}
	if len(build.Files()) != len(files) {
		t.Fatalf("files: %v", build.Files())
	}
}

// A file whose own recursion reads its summaries from the snapshot, and a
// file that cannot be read, keep incremental updates equal to full builds.
func TestFlowBuildUpdateSelfRecursionAndUnreadable(t *testing.T) {
	root := t.TempDir()
	opt := syntax.Options{Version: phpversion.Default}
	self := filepath.Join(root, "self.php")
	other := filepath.Join(root, "other.php")
	gone := filepath.Join(root, "gone.php")
	writeFile(t, self, "<?php function walk($n): void { if ($n) { walk($n - 1); } system($n); }")
	writeFile(t, other, "<?php function shell($c): void { exec($c); }")
	files := []string{gone, other, self}
	ix := BuildIndex(files, opt, infer.Options{})
	build := BuildFlow(context.Background(), files, nil, opt, ix, flow.Options{})
	for _, edit := range []struct{ path, src string }{
		{self, "<?php function walk($n): void { if ($n) { walk($n - 1); } echo $n; }"},
		{other, "<?php function shell($c): void { echo $c; }"},
	} {
		ix.Add(ExtractSymbols(edit.path, []byte(edit.src), opt, infer.Options{}))
		next, ok := build.Update(map[string][]byte{edit.path: []byte(edit.src)}, ix)
		if !ok {
			t.Fatal("update refused")
		}
		writeFile(t, edit.path, edit.src)
		if full := BuildFlow(context.Background(), files, nil, opt, ix, flow.Options{}); !next.Snapshot.Equal(full.Snapshot) {
			t.Fatalf("%s: incremental snapshot differs from a full build", filepath.Base(edit.path))
		}
		build = next
	}
}

// A chain of wrappers across files settles one file per round; past the
// rounds, the unsettled summaries are dropped and the call reads as
// unresolved rather than as clean.
func TestFlowChainAcrossFiles(t *testing.T) {
	for _, n := range []int{12, maxRounds + 1} {
		root := t.TempDir()
		writeFile(t, filepath.Join(root, "f00.php"), "<?php function f0() { return $_GET['v']; }")
		for i := 1; i < n; i++ {
			writeFile(t, filepath.Join(root, fmt.Sprintf("f%02d.php", i)), fmt.Sprintf("<?php function f%d() { return f%d(); }", i, i-1))
		}
		main := filepath.Join(root, "main.php")
		writeFile(t, main, fmt.Sprintf("<?php system(f%d());", n-1))
		p, err := Open([]analysis.Rule{untrustedshellcommand.New()}, Options{Root: root, Paths: []string{root}, Analysis: analysis.Config{EnableAll: true, PHP: phpversion.Default}, Parse: syntax.Options{Version: phpversion.Default}})
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range p.AnalyzeReport() {
			if r.Path != main {
				continue
			}
			settled := n <= maxRounds
			if got := len(r.Findings) == 1; got != settled {
				t.Errorf("chain of %d: findings %+v", n, r.Findings)
			}
			if got := len(r.Incomplete) == 1 && r.Incomplete[0].Reason == "unresolved-dispatch"; got == settled {
				t.Errorf("chain of %d: status %+v", n, r.Incomplete)
			}
		}
	}
}

func TestFlowBuildStopsWhenCancelled(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "a.php")
	writeFile(t, path, "<?php function source() { return $_GET['value']; }")
	opt := syntax.Options{Version: phpversion.Default}
	ix := BuildIndex([]string{path}, opt, infer.Options{})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if s := BuildFlowSnapshot(ctx, []string{path}, nil, opt, ix, flow.Options{}); s == nil || !s.Equal(flow.NewSnapshotWithOptions(flow.Options{})) {
		t.Fatal("a cancelled build summarized files")
	}
}

// A file whose summaries keep changing is summarized a bounded number of
// times, then dropped with what was computed from it.
func TestFlowOscillatingFileIsDropped(t *testing.T) {
	root := t.TempDir()
	flip := filepath.Join(root, "flip.php")
	reader := filepath.Join(root, "reader.php")
	writeFile(t, flip, "<?php function flip() { reader(); return $_GET['v']; }")
	writeFile(t, reader, "<?php function reader() { return flip(); }")
	opt := syntax.Options{Version: phpversion.Default}
	ix := BuildIndex([]string{flip, reader}, opt, infer.Options{})
	alternate := syntax.ParseBest(flip, []byte("<?php function flip() { reader(); return 1; }"), opt)
	old := extractFlow
	t.Cleanup(func() { extractFlow = old })
	var mu sync.Mutex
	calls := 0
	extractFlow = func(f *syntax.File, ix *index.Index, v phpversion.Version, s *flow.Snapshot, o flow.Options) *flow.File {
		if f.Path == flip {
			mu.Lock()
			calls++
			odd := calls%2 == 0
			mu.Unlock()
			if odd {
				f = alternate
			}
		}
		return old(f, ix, v, s, o)
	}
	build := BuildFlow(context.Background(), []string{flip, reader}, nil, opt, ix, flow.Options{})
	if calls != maxSummaries || build.results[0] != nil || build.results[1] != nil {
		t.Fatalf("summaries of flip: %d, kept %v", calls, build.results)
	}
}

// writeChain writes n files, each function returning the one before, the
// first returning request data.
func writeChain(t *testing.T, n int) []string {
	t.Helper()
	root := t.TempDir()
	var files []string
	for i := range n {
		body := "return $_GET['v'];"
		if i > 0 {
			body = fmt.Sprintf("return f%d();", i-1)
		}
		path := filepath.Join(root, fmt.Sprintf("f%02d.php", i))
		writeFile(t, path, fmt.Sprintf("<?php function f%d() { %s }", i, body))
		files = append(files, path)
	}
	return files
}

// An order from an earlier build summarizes each file of a chain once,
// whatever depths it gives, and builds the snapshot a build without one does.
func TestFlowBuildOrdered(t *testing.T) {
	files := writeChain(t, 6)
	opt := syntax.Options{Version: phpversion.Default}
	ix := BuildIndex(files, opt, infer.Options{})
	extracted := 0
	old := extractFlow
	t.Cleanup(func() { extractFlow = old })
	extractFlow = func(f *syntax.File, ix *index.Index, v phpversion.Version, s *flow.Snapshot, o flow.Options) *flow.File {
		extracted++
		return old(f, ix, v, s, o)
	}
	plain := BuildFlow(context.Background(), files, nil, opt, ix, flow.Options{})
	order := plain.Order()
	if order[files[0]] != 0 || order[files[5]] != 5 {
		t.Fatalf("order: %v", order)
	}
	extracted = 0
	ordered := BuildFlowOrdered(context.Background(), files, nil, opt, ix, flow.Options{}, order)
	if extracted != len(files) || !ordered.Snapshot.Equal(plain.Snapshot) {
		t.Fatalf("%d summaries", extracted)
	}
	// Past the waves, out of date or missing: only more summaries.
	stale := map[string]int{files[0]: 3 * maxWaves, files[1]: 2 * maxWaves}
	if b := BuildFlowOrdered(context.Background(), files, nil, opt, ix, flow.Options{}, stale); !b.Snapshot.Equal(plain.Snapshot) {
		t.Fatal("stale order")
	}
}

// A chain longer than the rounds loses its end without an order; one from
// an earlier build settles it whole.
func TestFlowBuildOrderedSettlesLongChains(t *testing.T) {
	files := writeChain(t, maxRounds+3)
	opt := syntax.Options{Version: phpversion.Default}
	ix := BuildIndex(files, opt, infer.Options{})
	plain := BuildFlow(context.Background(), files, nil, opt, ix, flow.Options{})
	if plain.results[maxRounds-1] == nil || plain.results[maxRounds] != nil {
		t.Fatal("the rounds bound the chain")
	}
	ordered := BuildFlowOrdered(context.Background(), files, nil, opt, ix, flow.Options{}, plain.Order())
	if slices.Contains(ordered.results, nil) {
		t.Fatal("an order settles the whole chain")
	}
}

// Files calling each other have the depth the first one visited gives.
func TestFlowOrderOfARecursion(t *testing.T) {
	root := t.TempDir()
	a, b := filepath.Join(root, "a.php"), filepath.Join(root, "b.php")
	writeFile(t, a, "<?php function a($n) { return $n ? b($n - 1) : $_GET['v']; }")
	writeFile(t, b, "<?php function b($n) { return a($n); }")
	opt := syntax.Options{Version: phpversion.Default}
	ix := BuildIndex([]string{a, b}, opt, infer.Options{})
	order := BuildFlow(context.Background(), []string{a, b}, nil, opt, ix, flow.Options{}).Order()
	if order[a] != 1 || order[b] != 0 {
		t.Fatalf("order: %v", order)
	}
}
