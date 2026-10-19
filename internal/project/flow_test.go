package project

import (
	"fmt"
	"path/filepath"
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
