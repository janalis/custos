package lsp

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/janalis/custos/internal/inspection/rules/possiblynullaccess"
	"github.com/janalis/custos/internal/project/baseline"
	"github.com/janalis/custos/internal/project/config"
	"github.com/janalis/custos/internal/testing/testbudget"
)

// actionsUntil requests code actions with params until ok accepts them and
// returns the last response. A background index or flow build publishes a
// new engine after the diagnostics a test waited for; until the document is
// analyzed again with it, the server answers with no actions, and a client
// asks again after the next publication.
func actionsUntil(t *testing.T, cl *client, params map[string]any, ok func([]codeAction) bool) *message {
	t.Helper()
	deadline := time.Now().Add(testbudget.Of(5 * time.Second))
	for id := 100; ; id++ {
		res := cl.call(id, "textDocument/codeAction", params)
		var actions []codeAction
		_ = json.Unmarshal(res.Result, &actions)
		if ok(actions) || time.Now().After(deadline) {
			return res
		}
		time.Sleep(10 * time.Millisecond)
	}
}

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
	// The shown one can be suppressed though the baseline hides the other.
	line := map[string]any{"line": 2, "character": 0}
	res := actionsUntil(t, cl, map[string]any{"textDocument": map[string]any{"uri": uri}, "range": map[string]any{"start": line, "end": line}, "context": map[string]any{"only": []string{"quickfix"}}}, func(a []codeAction) bool { return len(a) > 0 })
	if !strings.Contains(string(res.Result), "Suppress UnnecessarySemicolon") {
		t.Fatalf("actions: %s", res.Result)
	}
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

// TestBaselineShowsRepeatsOfRecordedFindings: PossiblyNullAccess reports
// the access after one the baseline records on the same value, which it
// leaves out after a reported one.
func TestBaselineShowsRepeatsOfRecordedFindings(t *testing.T) {
	dir := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(config.FileName, `{"baseline": "bl.json", "rules": {"PossiblyNullAccess": {"enabled": true}}}`)
	write("bl.json", `{"version": 1, "entries": [{"path": "a.php", "rule": "PossiblyNullAccess", "message": "Check this value for null before the access: it is Exception|null.", "line": "$a->getMessage();", "count": 1}]}`)
	useRules(t, possiblynullaccess.New())
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + dir})
	_ = cl.c.notify("initialized", map[string]any{})
	uri := "file://" + filepath.Join(dir, "a.php")
	cl.open(uri, "<?php\nfunction f(?Exception $a) {\n$a->getMessage();\n$a->getCode();\n$a->getLine();\n}\n")
	p := cl.diagsFor(uri)
	for len(p.Diagnostics) == 0 { // published again once the project is indexed
		p = cl.diagsFor(uri)
	}
	if len(p.Diagnostics) != 1 || p.Diagnostics[0].Range.Start.Line != 3 {
		t.Fatalf("diagnostics: %+v", p.Diagnostics)
	}
	// The shown finding, when no access follows it, can be suppressed: the
	// check of the edit analyzes with the baseline too.
	cl.change(uri, 2, "<?php\nfunction f(?Exception $a) {\n$a->getMessage();\n$a->getCode();\n}\n")
	// Diagnostics of the edited version: the first version, which has as
	// many, may be published again meanwhile (once the index or the flow
	// summaries are rebuilt), and its actions are gone.
	cl.diagsAt(uri, 2, 1)
	line := map[string]any{"line": 3, "character": 0}
	res := actionsUntil(t, cl, map[string]any{"textDocument": map[string]any{"uri": uri}, "range": map[string]any{"start": line, "end": line}, "context": map[string]any{}}, func(a []codeAction) bool { return len(a) > 0 })
	var actions []codeAction
	_ = json.Unmarshal(res.Result, &actions)
	if len(actions) != 1 || actions[0].Title != "Suppress PossiblyNullAccess for this statement" || actions[0].Edit == nil {
		t.Fatalf("actions: %s", res.Result)
	}
}
