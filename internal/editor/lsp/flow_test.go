package lsp

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"custos/internal/inspection/rules/untrustedshellcommand"
	"custos/internal/php/syntax"
)

func TestWorkspaceFlowRefreshesDiskAndUnsavedBuffer(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "wrapper.php")
	write := func(src string) {
		t.Helper()
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("<?php function invoke($arg) { system($arg); }")
	s := whiteBox(t, dir, io.Discard, untrustedshellcommand.New())
	if !s.beginIndexing() {
		t.Fatal("flow rule did not request the index")
	}
	s.buildIndex()
	check := func(want int) {
		t.Helper()
		f := syntax.Parse("main.php", []byte("<?php invoke($_GET['command']);"), syntax.Options{Version: s.cfg.PHP})
		if got := len(s.engine.Analyze(f)); got != want {
			t.Fatalf("wrapper diagnostics: %d, want %d", got, want)
		}
	}
	check(1)
	initial := s.flow
	uri := "file://" + path
	s.docs[uri] = &document{path: path, text: []byte("<?php function invoke($arg) { return $arg; }")}
	if !s.reindexDoc(uri) {
		t.Fatal("unsaved wrapper not indexed")
	}
	check(0)
	if s.flow == initial {
		t.Fatal("mutable flow snapshot publication")
	}
	delete(s.docs, uri)
	s.watchedFilesChanged([]fileChange{{URI: uri, Type: 2}})
	check(1)
	write("<?php function invoke($arg) { return $arg; }")
	s.watchedFilesChanged([]fileChange{{URI: uri, Type: 2}})
	check(0)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	s.watchedFilesChanged([]fileChange{{URI: uri, Type: 3}})
	check(0)
	// Failed discovery keeps the last published immutable snapshot.
	stable := s.flow
	s.cfg.Paths = []string{"missing"}
	s.refreshFlow()
	if s.flow != stable {
		t.Fatal("failed refresh replaced the snapshot")
	}
}

func TestWorkspaceFlowRefreshNeedsConfiguredIndex(t *testing.T) {
	(&Server{}).refreshFlow()
	s := whiteBox(t, t.TempDir(), io.Discard, untrustedshellcommand.New())
	s.refreshFlow()
	if s.flow != nil {
		t.Fatal("published flow without a project index")
	}
}

func TestInitialIndexPreservesPreviouslyOpenedFlowBuffers(t *testing.T) {
	unsafe := "<?php function invoke($arg) { system($arg); }"
	safe := "<?php function invoke($arg) { return $arg; }"
	for _, tc := range []struct {
		name, disk, buffer string
		count              int
	}{
		{"unsaved unsafe wrapper", safe, unsafe, 1},
		{"unsaved safe wrapper", unsafe, safe, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "wrapper.php")
			if err := os.WriteFile(path, []byte(tc.disk), 0o644); err != nil {
				t.Fatal(err)
			}
			s := whiteBox(t, dir, io.Discard, untrustedshellcommand.New())
			if !s.beginIndexing() {
				t.Fatal("flow inspection did not request initial indexing")
			}
			mainURI := "file://" + filepath.Join(dir, "main.php")
			s.open(didOpenParams{TextDocument: textDocumentItem{URI: mainURI, LanguageID: "php", Version: 1, Text: "<?php invoke($_GET['command']);"}})
			s.open(didOpenParams{TextDocument: textDocumentItem{URI: "file://" + path, LanguageID: "php", Version: 1, Text: tc.buffer}})
			if s.index != nil {
				t.Fatal("test opened buffers after initial index publication")
			}
			s.buildIndex()
			if s.indexing || s.index == nil || s.flow == nil {
				t.Fatal("initial index and flow snapshot were not published")
			}
			if got := len(s.docs[mainURI].findings); got != tc.count {
				t.Fatalf("initial publication used disk wrapper: got %d dependent findings, want %d", got, tc.count)
			}
			if s.docs[mainURI].analyzed != 1 {
				t.Fatal("initial index did not reanalyze the open dependent buffer")
			}
		})
	}
}
