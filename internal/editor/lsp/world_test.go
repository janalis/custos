package lsp

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"custos/internal/testing/testbudget"
)

// TestComposerChangeRebuildsWorld: a change to composer.json or to the
// installed.json of the vendor directory rebuilds the index, so its world
// follows the installed packages.
func TestComposerChangeRebuildsWorld(t *testing.T) {
	dir := t.TempDir()
	s := whiteBox(t, dir, io.Discard, unknownFuncRule{})
	if !s.beginIndexing() {
		t.Fatal("semantic rule enabled but no index wanted")
	}
	s.buildIndex()
	if s.index.World().Complete {
		t.Fatal("complete world without Composer metadata")
	}
	installed := filepath.Join(dir, "vendor", "composer", "installed.json")
	if err := os.MkdirAll(filepath.Dir(installed), 0o755); err != nil {
		t.Fatal(err)
	}
	for path, src := range map[string]string{filepath.Join(dir, "composer.json"): `{}`, installed: `{"packages": []}`} {
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	s.watchedFilesChanged([]fileChange{{URI: "file://" + installed, Type: 1}})
	deadline := time.Now().Add(testbudget.Of(5 * time.Second))
	for {
		s.mu.Lock()
		ix, busy := s.index, s.indexing
		s.mu.Unlock()
		if !busy && ix != nil && ix.World().Complete {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("world not recomputed")
		}
		time.Sleep(time.Millisecond)
	}
}

// TestComposerChangeOutsideWorld: a composer.json edit the world does not
// read (a description) keeps the index; one it reads rebuilds it.
func TestComposerChangeOutsideWorld(t *testing.T) {
	dir := t.TempDir()
	composer := filepath.Join(dir, "composer.json")
	write := func(src string) {
		t.Helper()
		if err := os.WriteFile(composer, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(`{"name": "acme/app", "autoload": {"psr-4": {"App\\": "src/"}}}`)
	s := whiteBox(t, dir, io.Discard, unknownFuncRule{})
	if !s.beginIndexing() {
		t.Fatal("semantic rule enabled but no index wanted")
	}
	s.buildIndex()
	generation := s.indexGeneration
	changed := []fileChange{{URI: "file://" + composer, Type: 2}}
	write(`{"autoload": {"psr-4": {"App\\": "src/"}}, "name": "acme/app", "description": "An app"}`)
	s.watchedFilesChanged(changed)
	if s.indexGeneration != generation {
		t.Fatal("composer.json edit outside the world rebuilt the index")
	}
	for _, src := range []string{
		`{"autoload": {"psr-4": {"App\\": "lib/"}}}`,
		`{"autoload": {"psr-4": {"App\\": "lib/"}}, "config": {"vendor-dir": "deps"}}`,
		`{`,
	} {
		write(src)
		s.watchedFilesChanged(changed)
		waitIndexed(t, s, generation+1)
		generation = s.indexGeneration
	}
}

// TestRebuildKeepsBuffers: an index rebuilt from disk indexes the open
// buffers again, flow rules on or not.
func TestRebuildKeepsBuffers(t *testing.T) {
	dir := t.TempDir()
	s := whiteBox(t, dir, io.Discard, unknownFuncRule{})
	lib := filepath.Join(dir, "lib.php")
	s.docs["file://"+lib] = &document{uri: "file://" + lib, path: lib, text: []byte("<?php\nfunction unsaved_helper() {}\n")}
	if !s.beginIndexing() {
		t.Fatal("semantic rule enabled but no index wanted")
	}
	s.buildIndex()
	if s.index.Function("unsaved_helper", s.cfg.PHP) == nil {
		t.Fatal("open buffer's declarations dropped by the rebuild")
	}
}

// waitIndexed waits for an index build of at least generation to end.
func waitIndexed(t *testing.T, s *Server, generation uint64) {
	t.Helper()
	deadline := time.Now().Add(testbudget.Of(5 * time.Second))
	for {
		s.mu.Lock()
		done := s.indexGeneration >= generation && !s.indexing
		s.mu.Unlock()
		if done {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("no index build %d", generation)
		}
		time.Sleep(time.Millisecond)
	}
}
