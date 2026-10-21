package mcp

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/inspection/analysis"
)

// The type tool infers under the workspace's flow options: with
// assertions trusted, an assert() narrows the value it reports.
func TestTypeHonorsAssertions(t *testing.T) {
	for _, trusted := range []bool{false, true} {
		root, storage := t.TempDir(), t.TempDir()
		if err := os.Chmod(storage, 0o700); err != nil {
			t.Fatal(err)
		}
		src := "<?php\nclass Code {}\nfunction f(?object $c) { assert($c instanceof Code); return $c; }\n"
		writeFile(t, filepath.Join(root, "a.php"), src)
		config := `{"php":"8.2"}`
		if trusted {
			config = `{"php":"8.2","flow":{"assertions":true}}`
		}
		writeFile(t, filepath.Join(root, "custos.json"), config)
		s, err := New(Options{Default: "default", Storage: storage, Workspaces: []Workspace{{ID: "default", Root: root, Readers: []string{"local"}, Writers: []string{"local"}}}}, []analysis.Rule{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(s.Close)
		ctx := context.WithValue(context.Background(), identityKey{}, "local")
		cp := invoke(ctx, t, s, "analyze", Input{})["checkpointId"].(string)
		got := invoke(ctx, t, s, "type", Input{CheckpointID: cp, Path: "a.php", Offset: uint32(strings.LastIndex(src, "$c;"))})["phpDocAware"]
		want := "null|object"
		if trusted {
			want = `\Code`
		}
		if got != want {
			t.Errorf("assertions %v: got %v want %s", trusted, got, want)
		}
	}
}
