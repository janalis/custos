package lsp

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"custos/internal/inspection/rules/unescapedhtmloutput"
	"custos/internal/php/syntax"
	"custos/internal/project"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

func TestIndexBuildComposesFlowOnce(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "source.php"), []byte("<?php function f($x) { return $x; }"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := whiteBox(t, root, io.Discard, unescapedhtmloutput.New())
	old := buildFlow
	defer func() { buildFlow = old }()
	builds := 0
	buildFlow = func(ctx context.Context, files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options flow.Options, order map[string]int) *project.FlowBuild {
		builds++
		return old(ctx, files, srcs, opt, ix, options, order)
	}
	if !s.beginIndexing() {
		t.Fatal("index not requested")
	}
	s.buildIndex()
	if builds != 1 || s.flow == nil {
		t.Fatalf("flow builds=%d snapshot=%v", builds, s.flow != nil)
	}
}

// TestFlowOrderSavedBetweenBuilds: with flow.order, a flow build saves its
// order and the next one summarizes in it; without a cache directory the
// builds go on without one.
func TestFlowOrderSavedBetweenBuilds(t *testing.T) {
	root := t.TempDir()
	for name, src := range map[string]string{
		"a.php": "<?php function a($x) { return b($x); }",
		"b.php": "<?php function b($x) { return $x; }",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("CUSTOS_CACHE_DIR", t.TempDir())
	s := whiteBox(t, root, io.Discard, unescapedhtmloutput.New())
	s.cfg.FlowOrder = true
	old := buildFlow
	defer func() { buildFlow = old }()
	var orders []map[string]int
	buildFlow = func(ctx context.Context, files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options flow.Options, order map[string]int) *project.FlowBuild {
		orders = append(orders, order)
		return old(ctx, files, srcs, opt, ix, options, order)
	}
	if !s.beginIndexing() {
		t.Fatal("index not requested")
	}
	s.buildIndex()
	s.refreshFlow()
	if len(orders) != 2 || len(orders[0]) != 0 || orders[1][filepath.Join(root, "a.php")] <= orders[1][filepath.Join(root, "b.php")] {
		t.Fatalf("orders: %v", orders)
	}
	t.Setenv("CUSTOS_CACHE_DIR", "")
	t.Setenv("HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	s.refreshFlow()
	if len(orders) != 3 || orders[2] != nil || s.flow == nil {
		t.Fatalf("orders without a cache directory: %v", orders)
	}
}
