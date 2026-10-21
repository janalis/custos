package lsp

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"custos/internal/project/baseline"
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

// TestBaselineReloadYieldsToReconfiguration: a baseline read while the
// configuration was replaced does not overwrite the new one's baseline.
func TestBaselineReloadYieldsToReconfiguration(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "bl.json"), []byte(`{"version": 1, "entries": []}`), 0o644); err != nil {
		t.Fatal(err)
	}
	s := whiteBox(t, dir, io.Discard)
	s.cfg = &config.Config{Root: dir, Baseline: "bl.json"}
	replaced := &config.Config{Root: dir}
	defer func(f func(string) (*baseline.Baseline, error)) { loadBaselineFile = f }(loadBaselineFile)
	loadBaselineFile = func(path string) (*baseline.Baseline, error) {
		s.mu.Lock()
		s.cfg = replaced
		s.mu.Unlock()
		return baseline.Load(path)
	}
	s.reloadBaseline()
	if s.baseline != nil || s.cfg != replaced {
		t.Fatalf("stale baseline installed: %+v", s.baseline)
	}
}

// TestBaselineWithoutConfiguration: file changes reported after the
// project's configuration failed to load do not look for a baseline.
func TestBaselineWithoutConfiguration(t *testing.T) {
	s := whiteBox(t, t.TempDir(), io.Discard)
	s.cfg = nil
	s.watchedFilesChanged([]fileChange{{URI: "file:///w/bl.json", Type: 2}})
	if baselinePath(nil) != "" || s.baseline != nil {
		t.Fatal("baseline without configuration")
	}
}
