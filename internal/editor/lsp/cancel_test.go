package lsp

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/janalis/custos/internal/inspection/rules/untrustedshellcommand"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/project"
	"github.com/janalis/custos/internal/semantic/flow"
	"github.com/janalis/custos/internal/semantic/index"
)

// supersede ends the flow work in progress, as a later full build does.
func (s *Server) supersede() {
	s.mu.Lock()
	s.flowCancel()
	s.mu.Unlock()
}

// Flow work a later build supersedes is dropped, not published: it read
// older buffers and may have stopped half way. A new index build stops
// the one in progress.
func TestSupersededWorkIsDropped(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "source.php")
	source := []byte(`<?php function local($x){system($x);}`)
	if err := os.WriteFile(path, source, 0o644); err != nil {
		t.Fatal(err)
	}
	s := whiteBox(t, root, io.Discard, untrustedshellcommand.New())
	if !s.beginIndexing() {
		t.Fatal("index not requested")
	}
	first := s.indexCtx
	if !s.beginIndexing() || first.Err() == nil {
		t.Fatal("a new index build must stop the one in progress")
	}
	s.buildIndex()
	published := s.flowBuild
	if published == nil {
		t.Fatal("no flow build")
	}

	s.setFlowBuilder(func(ctx context.Context, files []string, sources [][]byte, opt syntax.Options, ix *index.Index, options flow.Options, order map[string]int) *project.FlowBuild {
		build := project.BuildFlowOrdered(ctx, files, sources, opt, ix, options, order)
		s.supersede()
		return build
	})
	defer s.setFlowBuilder(nil)
	s.refreshFlow()
	if s.flowBuild != published {
		t.Fatal("a superseded build was published")
	}

	uri := "file://" + path
	s.docs[uri] = &document{path: path, text: append(append([]byte(nil), source...), " // edited"...)}
	s.setFlowUpdater(func(ctx context.Context, b *project.FlowBuild, changed map[string][]byte, ix *index.Index) (*project.FlowBuild, bool) {
		s.supersede()
		return b.Update(ctx, changed, ix)
	})
	defer s.setFlowUpdater(nil)
	if s.reindexDoc(uri) || s.flowBuild != published {
		t.Fatal("a superseded update was published")
	}
}
