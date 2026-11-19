package lsp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/pattern"
)

// A project's custom rules report diagnostics and their fix-rule command
// applies only their quick-fixes.
func TestCustomRuleCommand(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "custos.json"), []byte(`{"customRules":[{"id":"NoTriple","match":"!!!$X","message":"m","fix":"!$X"}]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + root, "initializationOptions": map[string]any{"rules": map[string]any{"NoTriple": map[string]any{"severity": "error"}}}})
	uri := "file://" + filepath.Join(root, "a.php")
	cl.open(uri, "<?php\n$a = !!!$b;;\n")
	if d := cl.diagsFor(uri); !strings.Contains(mustMarshal(t, d), `"code":"NoTriple"`) || !strings.Contains(mustMarshal(t, d), `"severity":1`) {
		t.Fatalf("diagnostics: %s", mustMarshal(t, d))
	}
	raw := json.RawMessage(mustString(50))
	_ = cl.c.write(&message{ID: &raw, Method: "workspace/executeCommand", Params: mustJSON(map[string]any{"command": cmdFixRule, "arguments": []any{uri, "NoTriple"}})})
	m := cl.waitFor("applyEdit", func(m *message) bool { return m.Method == "workspace/applyEdit" })
	if !strings.Contains(string(m.Params), `?php\n$a = !$b;;\n"`) {
		t.Fatalf("edit: %s", m.Params)
	}
}

// Rules added by a program embedding custos run and fix as the others do.
func TestExtraRuleCommand(t *testing.T) {
	extra, err := pattern.Compile(pattern.Definition{ID: "NoTriple", Match: pattern.Patterns{"!!!$X"}, Message: "m", Fix: &[]string{"!$X"}[0]})
	if err != nil {
		t.Fatal(err)
	}
	cl := startServerWith(t, []analysis.Rule{extra})
	cl.initialize(map[string]any{"rootUri": "file://" + t.TempDir()})
	uri := "file:///tmp/extra.php"
	cl.open(uri, "<?php\n$a = !!!$b;\n")
	if d := cl.diagsFor(uri); !strings.Contains(mustMarshal(t, d), `"code":"NoTriple"`) {
		t.Fatalf("diagnostics: %s", mustMarshal(t, d))
	}
	raw := json.RawMessage(mustString(51))
	_ = cl.c.write(&message{ID: &raw, Method: "workspace/executeCommand", Params: mustJSON(map[string]any{"command": cmdFixRule, "arguments": []any{uri, "NoTriple"}})})
	m := cl.waitFor("applyEdit", func(m *message) bool { return m.Method == "workspace/applyEdit" })
	if !strings.Contains(string(m.Params), `?php\n$a = !$b;\n"`) {
		t.Fatalf("edit: %s", m.Params)
	}
}

func mustMarshal(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// Edits to custos.json and to its rule files take effect without a
// restart; a broken configuration is reported and its fix picked up.
func TestCustomRuleReload(t *testing.T) {
	dir := t.TempDir()
	write := func(name, text string) {
		t.Helper()
		if err := os.WriteFile(filepath.Join(dir, name), []byte(text), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("custos.json", `{"customRuleFiles": ["rules.json"]}`)
	write("rules.json", `[{"id": "NoTriple", "match": "!!!$X", "message": "first"}]`)
	cl := startServer(t)
	cl.initialize(map[string]any{
		"rootUri":      "file://" + dir,
		"capabilities": map[string]any{"workspace": map[string]any{"didChangeWatchedFiles": map[string]any{"dynamicRegistration": true}}},
	})
	_ = cl.c.notify("initialized", map[string]any{})
	if reg := cl.wait("client/registerCapability"); !strings.Contains(string(reg.Params), "**/custos.json") || !strings.Contains(string(reg.Params), "**/rules.json") {
		t.Fatalf("registration: %s", reg.Params)
	}
	uri := "file://" + filepath.Join(dir, "a.php")
	cl.open(uri, "<?php\n$a = !!!$b;\n")
	if d := mustMarshal(t, cl.diagsFor(uri)); !strings.Contains(d, `"message":"first"`) {
		t.Fatalf("diagnostics: %s", d)
	}
	changed := func(name string) {
		_ = cl.c.notify("workspace/didChangeWatchedFiles", map[string]any{"changes": []any{map[string]any{"uri": "file://" + filepath.Join(dir, name), "type": 2}}})
	}
	write("rules.json", `[{"id": "NoTriple", "match": "!!!$X", "message": "second"}]`)
	changed("rules.json")
	cl.waitFor("diagnostics after rule edit", func(m *message) bool {
		return m.Method == "textDocument/publishDiagnostics" && strings.Contains(string(m.Params), `"message":"second"`)
	})
	write("custos.json", `{"customRules": [{"id": "Bad"}]}`)
	changed("custos.json")
	cl.logContaining("configuration:")
	write("custos.json", `{"customRules": [{"id": "NoTriple", "match": "!!!$X", "message": "third"}]}`)
	changed("custos.json")
	cl.waitFor("diagnostics after config edit", func(m *message) bool {
		return m.Method == "textDocument/publishDiagnostics" && strings.Contains(string(m.Params), `"message":"third"`)
	})
}
