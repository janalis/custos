package lsp

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
	"custos/internal/project"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

func (s *Server) setFlowBuilder(builder flowBuilder) {
	s.mu.Lock()
	s.flowBuilder = builder
	s.mu.Unlock()
}

func (s *Server) setFlowUpdater(updater flowUpdater) {
	s.mu.Lock()
	s.flowUpdater = updater
	s.mu.Unlock()
}

func TestFlowHooksStayWithTheirServer(t *testing.T) {
	source := []byte(`<?php function local($x){system($x);}`)
	makeServer := func() (*Server, string, string) {
		root := t.TempDir()
		path := filepath.Join(root, "source.php")
		if err := os.WriteFile(path, source, 0o644); err != nil {
			t.Fatal(err)
		}
		return whiteBox(t, root, io.Discard, untrustedshellcommand.New()), path, "file://" + path
	}
	left, leftPath, leftURI := makeServer()
	right, rightPath, rightURI := makeServer()
	builds, updates := 0, 0
	left.setFlowBuilder(func(ctx context.Context, files []string, sources [][]byte, opt syntax.Options, ix *index.Index, options flow.Options, order map[string]int) *project.FlowBuild {
		builds++
		return project.BuildFlowOrdered(ctx, files, sources, opt, ix, options, order)
	})
	defer left.setFlowBuilder(nil)
	for _, server := range []*Server{left, right} {
		if !server.beginIndexing() {
			t.Fatal("index not requested")
		}
		server.buildIndex()
	}
	if builds != 1 {
		t.Fatal("another server used the build override")
	}
	left.setFlowUpdater(func(build *project.FlowBuild, changed map[string][]byte, ix *index.Index) (*project.FlowBuild, bool) {
		updates++
		return build.Update(changed, ix)
	})
	defer left.setFlowUpdater(nil)
	left.docs[leftURI] = &document{path: leftPath, text: append(append([]byte(nil), source...), []byte(" // edited")...)}
	right.docs[rightURI] = &document{path: rightPath, text: append(append([]byte(nil), source...), []byte(" // edited")...)}
	right.reindexDoc(rightURI)
	if updates != 0 {
		t.Fatal("another server used the update override")
	}
	left.reindexDoc(leftURI)
	if updates != 1 {
		t.Fatal("the owning server lost its update override")
	}
}
