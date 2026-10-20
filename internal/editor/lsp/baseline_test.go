package lsp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/project/config"
)

// TestBaselineHidesKnownFindings: the project baseline hides the findings
// it records, matched on the open buffer's line text, and follows the
// baseline file when it changes on disk.
func TestBaselineHidesKnownFindings(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	entry := func(count string) string {
		return `{"version": 1, "entries": [{"path": "a.php", "rule": "UnnecessarySemicolon", "message": "empty statement", "line": ";", "count": ` + count + `}]}`
	}
	write(config.FileName, `{"baseline": "bl.json"}`)
	write("bl.json", entry("1"))
	cl := startServer(t)
	cl.initialize(map[string]any{
		"rootUri":      "file://" + dir,
		"capabilities": map[string]any{"workspace": map[string]any{"didChangeWatchedFiles": map[string]any{"dynamicRegistration": true}}},
	})
	_ = cl.c.notify("initialized", map[string]any{})
	if reg := cl.wait("client/registerCapability"); !strings.Contains(string(reg.Params), "**/bl.json") {
		t.Fatalf("registration: %s", reg.Params)
	}
	uri := "file://" + filepath.Join(dir, "a.php")
	cl.open(uri, "<?php\n;\n")
	cl.diagsN(uri, 0)
	// A second occurrence of the same line exceeds the recorded count.
	cl.change(uri, 2, "<?php\n ;\n;\n")
	cl.diagsN(uri, 1)
	changed := map[string]any{"changes": []any{map[string]any{"uri": "file://" + filepath.Join(dir, "bl.json"), "type": 2}}}
	write("bl.json", entry("2"))
	_ = cl.c.notify("workspace/didChangeWatchedFiles", changed)
	cl.diagsN(uri, 0)
	// An unreadable baseline is reported and shows every finding.
	write("bl.json", "{")
	_ = cl.c.notify("workspace/didChangeWatchedFiles", changed)
	cl.logContaining("baseline:")
	cl.diagsN(uri, 2)
}
