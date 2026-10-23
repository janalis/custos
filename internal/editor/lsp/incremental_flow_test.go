package lsp

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
	"custos/internal/project"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

// An edit that keeps a buffer's declarations updates the flow build: other
// documents are analysed again only when a contract changed, and the
// summaries match a full build. Any other edit rebuilds them.
func TestIncrementalFlowRefresh(t *testing.T) {
	root := t.TempDir()
	s := whiteBox(t, root, io.Discard, untrustedshellcommand.New())
	wrapper := filepath.Join(root, "wrapper.php")
	if err := os.WriteFile(wrapper, []byte("<?php function invoke($arg) { system($arg); }"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !s.beginIndexing() {
		t.Fatal("index not requested")
	}
	s.buildIndex()
	uri := "file://" + wrapper
	s.docs[uri] = &document{path: wrapper, text: []byte("<?php function invoke($arg) { system($arg); }")}
	findings := func() int {
		f := syntax.Parse("main.php", []byte("<?php invoke($_GET['c']);"), syntax.Options{Version: s.cfg.PHP})
		return len(s.engine.Analyze(f))
	}
	edit := func(src string) bool {
		t.Helper()
		s.docs[uri].text = []byte(src)
		return s.reindexDoc(uri)
	}
	// The buffer holds the file as the build read it: nothing to redo.
	opened := s.flowBuild
	if edit("<?php function invoke($arg) { system($arg); }") || s.flowBuild != opened || s.flowSymbols[wrapper] == nil {
		t.Fatal("opening a saved file summarized it again")
	}
	full := s.flowBuild
	if edit("<?php function invoke($arg) { system($arg); /* typing */ }") {
		t.Fatal("an edit changing no contract reanalysed every document")
	}
	if s.flowBuild == full || findings() != 1 {
		t.Fatal("unchanged contract: build not updated or findings lost")
	}
	engine := s.engine
	if !edit("<?php function invoke($arg) { echo $arg; }") || s.engine == engine || findings() != 0 {
		t.Fatal("removing the sink did not republish the summaries")
	}
	sameAsFull := func() {
		t.Helper()
		files := s.flowBuild.Files()
		srcs := make([][]byte, len(files))
		for i, f := range files {
			if f == wrapper {
				srcs[i] = s.docs[uri].text
			}
		}
		if !project.BuildFlow(context.Background(), files, srcs, syntax.Options{Version: s.cfg.PHP}, s.index, s.cfg.Flow).Snapshot.Equal(s.flow) {
			t.Fatal("incremental summaries differ from a full build")
		}
	}
	sameAsFull()
	// A changed declaration rebuilds.
	build := s.flowBuild
	if !edit("<?php function invoke($arg, $more = null) { system($arg); }") || s.flowBuild == build || findings() != 1 {
		t.Fatal("declaration change not rebuilt")
	}
	sameAsFull()
	// Another refresh published meanwhile: the update is dropped and the
	// summaries rebuilt.
	defer func(f func(*project.FlowBuild, map[string][]byte, *index.Index) (*project.FlowBuild, bool)) {
		updateFlow = f
	}(updateFlow)
	updateFlow = func(b *project.FlowBuild, changed map[string][]byte, ix *index.Index) (*project.FlowBuild, bool) {
		next, ok := b.Update(changed, ix)
		s.flowBuild = nil
		return next, ok
	}
	if !edit("<?php function invoke($arg, $more = null) { system($arg); // concurrent\n}") || s.flowBuild == nil {
		t.Fatal("superseded update not rebuilt")
	}
	updateFlow = func(*project.FlowBuild, map[string][]byte, *index.Index) (*project.FlowBuild, bool) {
		return nil, false
	}
	if !edit("<?php function invoke($arg, $more = null) { system($arg); // refused\n}") {
		t.Fatal("refused update not rebuilt")
	}
}

// A document changed while its analysis waited for a slot is not analysed
// for the superseded text.
func TestAnalysisSkippedWhenSupersededWhileWaiting(t *testing.T) {
	var out bytes.Buffer
	s := whiteBox(t, t.TempDir(), &out, untrustedshellcommand.New())
	uri := "file:///w/a.php"
	s.docs[uri] = &document{path: "/w/a.php", text: []byte("<?php system($_GET['c']);"), lines: syntax.NewLineIndex([]byte("<?php system($_GET['c']);")), analyzed: -1}
	defer func(f func(*Server)) { acquireSlot = f }(acquireSlot)
	analysed := false
	acquireSlot = func(s *Server) {
		s.docs[uri].revision++
		s.sem <- struct{}{}
	}
	s.analyzeNow(uri)
	if s.docs[uri].analyzed != -1 || strings.Contains(out.String(), "publishDiagnostics") {
		t.Fatal("superseded analysis published")
	}
	acquireSlot = func(s *Server) { analysed = true; s.sem <- struct{}{} }
	s.analyzeNow(uri)
	if !analysed || !strings.Contains(out.String(), "publishDiagnostics") || len(s.sem) != 0 {
		t.Fatal("current analysis not published or slot kept")
	}
}

// An edited document is analysed before its summaries are updated, and the
// updates run one at a time: an edit arriving meanwhile is summarized once
// the running update ends, and a failed update does not stop the next.
func TestFlowUpdateAfterAnalysis(t *testing.T) {
	root := t.TempDir()
	var out bytes.Buffer
	s := whiteBox(t, root, &out, untrustedshellcommand.New())
	wrapper := filepath.Join(root, "wrapper.php")
	src := []byte("<?php function invoke($arg) { system($arg); }")
	if err := os.WriteFile(wrapper, src, 0o644); err != nil {
		t.Fatal(err)
	}
	if !s.beginIndexing() {
		t.Fatal("index not requested")
	}
	s.buildIndex()
	uri := "file://" + wrapper
	edit := func(text string) {
		s.mu.Lock()
		s.docs[uri] = &document{uri: uri, path: wrapper, text: []byte(text), lines: syntax.NewLineIndex([]byte(text)), analyzed: -1}
		s.mu.Unlock()
	}
	edit(string(src))
	s.reflow(uri) // records the buffer's symbols
	defer func(f func(*project.FlowBuild, map[string][]byte, *index.Index) (*project.FlowBuild, bool)) {
		updateFlow = f
	}(updateFlow)
	var updates []string
	running := false
	updateFlow = func(b *project.FlowBuild, changed map[string][]byte, ix *index.Index) (*project.FlowBuild, bool) {
		if !strings.Contains(out.String(), "publishDiagnostics") {
			t.Error("summaries updated before the document was analysed")
		}
		if running {
			t.Error("two updates ran at once")
		}
		running = true
		defer func() { running = false }()
		updates = append(updates, string(changed[wrapper]))
		if len(updates) == 1 {
			edit("<?php function invoke($arg) { system($arg); /* later */ }")
			s.reflow(uri) // queued behind the running update
		}
		return b.Update(changed, ix)
	}
	edit("<?php function invoke($arg) { system($arg); /* typing */ }")
	out.Reset()
	s.analyzeChanged(uri)
	if len(updates) != 2 || !strings.Contains(updates[1], "later") || s.reflowing {
		t.Fatalf("updates: %q reflowing=%v", updates, s.reflowing)
	}
	updateFlow = func(*project.FlowBuild, map[string][]byte, *index.Index) (*project.FlowBuild, bool) {
		panic("update")
	}
	s.reflow(uri)
	if s.reflowing || !strings.Contains(out.String(), "flow update panicked") {
		t.Fatal("failed update left the queue busy or unreported")
	}
}

// A document opened with other text than the saved file updates the build
// the file was summarized in when the declarations are the same, and
// rebuilds it otherwise or when the file cannot be read.
func TestOpenedDocumentUpdatesBuild(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		remove     bool
		rebuilt    bool
	}{
		{"same declarations", "<?php function invoke($arg) { echo $arg; }", false, false},
		{"other declarations", "<?php function invoke($arg, $more = null) { system($arg); }", false, true},
		{"unreadable file", "<?php function invoke($arg) { echo $arg; }", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			s := whiteBox(t, root, io.Discard, untrustedshellcommand.New())
			wrapper := filepath.Join(root, "wrapper.php")
			if err := os.WriteFile(wrapper, []byte("<?php function invoke($arg) { system($arg); }"), 0o644); err != nil {
				t.Fatal(err)
			}
			if !s.beginIndexing() {
				t.Fatal("index not requested")
			}
			s.buildIndex()
			if tc.remove {
				if err := os.Remove(wrapper); err != nil {
					t.Fatal(err)
				}
			}
			old := buildFlow
			defer func() { buildFlow = old }()
			builds := 0
			buildFlow = func(ctx context.Context, files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options flow.Options, order map[string]int) *project.FlowBuild {
				builds++
				return old(ctx, files, srcs, opt, ix, options, order)
			}
			uri := "file://" + wrapper
			s.docs[uri] = &document{path: wrapper, text: []byte(tc.text)}
			if !s.reindexDoc(uri) || (builds > 0) != tc.rebuilt {
				t.Fatalf("builds=%d", builds)
			}
		})
	}
}
