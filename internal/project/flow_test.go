package project

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
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
	ix := BuildIndex([]string{path}, opt)
	fromDisk := BuildFlowSnapshot([]string{path, filepath.Join(root, "gone.php")}, nil, opt, ix)
	fromBuffer := BuildFlowSnapshot([]string{path}, [][]byte{[]byte("<?php function source() { return $_GET['value']; }")}, opt, ix)
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
	ix.Add(ExtractSymbols(path, []byte(chain.String()), opt))
	long := BuildFlowSnapshot([]string{path}, [][]byte{[]byte(chain.String())}, opt, ix)
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
		ix.Add(ExtractSymbols(paths[i], srcs[i], opt))
	}
	if BuildFlowSnapshot(paths, srcs, opt, ix) == nil {
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
	ix := BuildIndex(files, opt)
	build := BuildFlow(files, nil, opt, ix)
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
		ix.Add(ExtractSymbols(path, []byte(edit.src), opt))
		next, ok := build.Update(map[string][]byte{path: []byte(edit.src)}, ix)
		if !ok {
			t.Fatalf("step %d: update refused", step)
		}
		srcs := make([][]byte, len(files))
		for i, f := range files {
			srcs[i] = []byte(current[f])
		}
		if full := BuildFlow(files, srcs, opt, ix); !next.Snapshot.Equal(full.Snapshot) {
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
	ix := BuildIndex(files, opt)
	build := BuildFlow(files, nil, opt, ix)
	for _, edit := range []struct{ path, src string }{
		{self, "<?php function walk($n): void { if ($n) { walk($n - 1); } echo $n; }"},
		{other, "<?php function shell($c): void { echo $c; }"},
	} {
		ix.Add(ExtractSymbols(edit.path, []byte(edit.src), opt))
		next, ok := build.Update(map[string][]byte{edit.path: []byte(edit.src)}, ix)
		if !ok {
			t.Fatal("update refused")
		}
		writeFile(t, edit.path, edit.src)
		if full := BuildFlow(files, nil, opt, ix); !next.Snapshot.Equal(full.Snapshot) {
			t.Fatalf("%s: incremental snapshot differs from a full build", filepath.Base(edit.path))
		}
		build = next
	}
}
