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
