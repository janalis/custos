package lsp

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/infer"
)

// TestWorkspaceIndexUsesCache: a second index build of an unchanged
// workspace reads its declarations from the cache.
func TestWorkspaceIndexUsesCache(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "Lib.php")
	if err := os.WriteFile(path, []byte("<?php class Lib {}"), 0o644); err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(path, past, past); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CUSTOS_CACHE_DIR", t.TempDir())
	opt := syntax.Options{Version: phpversion.Default}
	for range 2 {
		if indexWorkspace(root, []string{path}, opt, infer.Options{}).Class("Lib", phpversion.Default) == nil {
			t.Fatal("class not indexed")
		}
	}
	entries, err := os.ReadDir(os.Getenv("CUSTOS_CACHE_DIR"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("cache files: %v %v", entries, err)
	}
	t.Setenv("CUSTOS_CACHE_DIR", "")
	t.Setenv("HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	if indexWorkspace(root, []string{path}, opt, infer.Options{}).Class("Lib", phpversion.Default) == nil {
		t.Fatal("index without a cache directory")
	}
}
