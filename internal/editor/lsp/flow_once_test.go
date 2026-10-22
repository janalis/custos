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
	buildFlow = func(ctx context.Context, files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options flow.Options) *project.FlowBuild {
		builds++
		return old(ctx, files, srcs, opt, ix, options)
	}
	if !s.beginIndexing() {
		t.Fatal("index not requested")
	}
	s.buildIndex()
	if builds != 1 || s.flow == nil {
		t.Fatalf("flow builds=%d snapshot=%v", builds, s.flow != nil)
	}
}
