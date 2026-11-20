package lsp

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/janalis/custos/internal/format"
	"github.com/janalis/custos/internal/inspection/rules/formatting"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/testing/testbudget"
)

const unformatted = "<?php\n\nfunction total($a,$b){\n  return $a+$b;\n}\n\nfunction other($c){\n  return $c;\n}\n"

// formatEdits sends a formatting request and returns its edits.
func (cl *client) formatEdits(id int, method string, params map[string]any) []textEdit {
	cl.t.Helper()
	res := cl.call(id, method, params)
	if res.Error != nil {
		cl.t.Fatalf("%s: %v", method, res.Error)
	}
	var edits []textEdit
	if err := json.Unmarshal(res.Result, &edits); err != nil || edits == nil {
		cl.t.Fatalf("%s: result %s (%v)", method, res.Result, err)
	}
	return edits
}

// applyEdits applies LSP edits to text, as a client does.
func applyEdits(text string, edits []textEdit) string {
	lines := syntax.NewLineIndex([]byte(text))
	type span struct {
		start, end uint32
		text       string
	}
	var spans []span
	for _, e := range edits {
		spans = append(spans, span{lines.OffsetUTF16(e.Range.Start.Line, e.Range.Start.Character), lines.OffsetUTF16(e.Range.End.Line, e.Range.End.Character), e.NewText})
	}
	slices.SortStableFunc(spans, func(a, b span) int { return int(b.start) - int(a.start) })
	for _, s := range spans {
		text = text[:s.start] + s.text + text[s.end:]
	}
	return text
}

func TestFormatting(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"composer.json":        "{}",
		"custos.json":          `{"format": {"exclude": ["legacy"]}}`,
		".editorconfig":        "root = true\n",
		"odd/.editorconfig/.x": "",
	})
	cl := startServer(t)
	res := cl.call(1, "initialize", map[string]any{"rootUri": "file://" + dir, "capabilities": map[string]any{}})
	if res.Error != nil || !strings.Contains(string(res.Result), `"documentFormattingProvider":true`) || !strings.Contains(string(res.Result), `"documentRangeFormattingProvider":true`) {
		t.Fatalf("capabilities: %s %v", res.Result, res.Error)
	}
	uri := "file://" + filepath.Join(dir, "a.php")
	cl.open(uri, unformatted)
	doc := map[string]any{"uri": uri}
	want := string(format.Source("a.php", []byte(unformatted), format.Options{}).Output)
	if got := applyEdits(unformatted, cl.formatEdits(2, "textDocument/formatting", map[string]any{"textDocument": doc, "options": map[string]any{"tabSize": 2, "insertSpaces": false}})); got != want {
		t.Fatalf("formatting gave\n%s\nwant\n%s", got, want)
	}
	// Lines 7-9 (0-based 6-8) hold the second function only.
	second := applyEdits(unformatted, cl.formatEdits(3, "textDocument/rangeFormatting", map[string]any{"textDocument": doc, "range": map[string]any{"start": map[string]any{"line": 6, "character": 3}, "end": map[string]any{"line": 9, "character": 0}}}))
	if !strings.Contains(second, "function total($a,$b){") || !strings.Contains(second, "function other($c)\n{\n    return $c;\n}") {
		t.Fatalf("range formatting gave\n%s", second)
	}
	// A reversed range, and one past the end, format nothing outside it.
	if edits := cl.formatEdits(4, "textDocument/rangeFormatting", map[string]any{"textDocument": doc, "range": map[string]any{"start": map[string]any{"line": 1, "character": 0}, "end": map[string]any{"line": 0, "character": 2}}}); len(edits) != 0 {
		t.Fatalf("lines 1-2 hold nothing to format: %+v", edits)
	}
	if edits := cl.formatEdits(5, "textDocument/rangeFormatting", map[string]any{"textDocument": doc, "range": map[string]any{"start": map[string]any{"line": 40, "character": 0}, "end": map[string]any{"line": 41, "character": 0}}}); len(edits) != 0 {
		t.Fatalf("past the end: %+v", edits)
	}
	// Nothing for a file left out, unknown, broken or whose settings fail.
	for i, tc := range []struct{ name, text string }{
		{"legacy/a.php", unformatted},
		{"b.php", "<?php\nfunction f($a,$b){ return $a +; }\n"},
		{"odd/a.php", unformatted},
	} {
		other := "file://" + filepath.Join(dir, tc.name)
		cl.open(other, tc.text)
		if edits := cl.formatEdits(10+i, "textDocument/formatting", map[string]any{"textDocument": map[string]any{"uri": other}}); len(edits) != 0 {
			t.Errorf("%s: %+v", tc.name, edits)
		}
	}
	if edits := cl.formatEdits(20, "textDocument/formatting", map[string]any{"textDocument": map[string]any{"uri": "file:///nowhere.php"}}); len(edits) != 0 {
		t.Fatalf("unknown document: %+v", edits)
	}
	for i, params := range []any{"bad", map[string]any{"textDocument": doc}} {
		if res := cl.call(30+i, "textDocument/rangeFormatting", params); res.Error == nil || res.Error.Code != codeInvalidParams {
			t.Fatalf("invalid params accepted: %+v", res)
		}
	}
}

// Editing an .editorconfig file changes what the Formatting rule reports.
func TestFormattingRuleFollowsEditorConfig(t *testing.T) {
	useRules(t, formatting.New())
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{
		"composer.json": "{}",
		"custos.json":   `{"rules": {"Formatting": {"enabled": true}}}`,
		".editorconfig": "root = true\n",
	})
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + dir, "capabilities": map[string]any{"workspace": map[string]any{"didChangeWatchedFiles": map[string]any{"dynamicRegistration": true}}}})
	_ = cl.c.notify("initialized", map[string]any{})
	if reg := cl.wait("client/registerCapability"); !strings.Contains(string(reg.Params), "**/.editorconfig") {
		t.Fatalf("registration: %s", reg.Params)
	}
	uri := "file://" + filepath.Join(dir, "a.php")
	cl.open(uri, "<?php\n\nfunction f()\n{\n  return 1;\n}\n")
	cl.diagsN(uri, 1)
	ec := filepath.Join(dir, ".editorconfig")
	if err := os.WriteFile(ec, []byte("root = true\n\n[*.php]\nindent_size = 2\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	other := filepath.Join(dir, "b.php")
	if err := os.WriteFile(other, []byte("<?php\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = cl.c.notify("workspace/didChangeWatchedFiles", map[string]any{"changes": []any{map[string]any{"uri": "file://" + other, "type": 1}}})
	_ = cl.c.notify("workspace/didChangeWatchedFiles", map[string]any{"changes": []any{map[string]any{"uri": "file://" + ec, "type": 2}}})
	cl.diagsN(uri, 0)
}

// writeTree writes files (name → content) under dir.
func writeTree(t *testing.T, dir string, files map[string]string) {
	t.Helper()
	for name, body := range files {
		p := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

// Editing pint.json, which the auto preset imports, changes what
// formatting writes.
func TestFormattingFollowsPint(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{"composer.json": "{}", "pint.json": `{"preset": "per"}`})
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + dir, "capabilities": map[string]any{}})
	_ = cl.c.notify("initialized", map[string]any{})
	uri := "file://" + filepath.Join(dir, "a.php")
	src := "<?php\n\n$a = $b.$c;\n"
	cl.open(uri, src)
	doc := map[string]any{"uri": uri}
	if got := applyEdits(src, cl.formatEdits(2, "textDocument/formatting", map[string]any{"textDocument": doc})); !strings.Contains(got, "$b . $c") {
		t.Fatalf("per: %s", got)
	}
	pint := filepath.Join(dir, "pint.json")
	if err := os.WriteFile(pint, []byte(`{"preset": "laravel"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = cl.c.notify("workspace/didChangeWatchedFiles", map[string]any{"changes": []any{map[string]any{"uri": "file://" + pint, "type": 2}}})
	deadline := time.Now().Add(testbudget.Of(5 * time.Second))
	for id := 3; ; id++ {
		got := applyEdits(src, cl.formatEdits(id, "textDocument/formatting", map[string]any{"textDocument": doc}))
		if strings.Contains(got, "$b.$c") {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("laravel after the change: %s", got)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// A range holding one of the edits that only work together formats none
// of them.
func TestRangeFormattingKeepsGroups(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{"composer.json": "{}"})
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + dir, "capabilities": map[string]any{}})
	uri := "file://" + filepath.Join(dir, "a.php")
	src := "<?php\n\n$a = array(\n    1,\n);\n"
	cl.open(uri, src)
	doc := map[string]any{"uri": uri}
	line := func(n int) map[string]any { return map[string]any{"line": n, "character": 0} }
	got := applyEdits(src, cl.formatEdits(2, "textDocument/rangeFormatting", map[string]any{"textDocument": doc, "range": map[string]any{"start": line(2), "end": line(3)}}))
	if got != src {
		t.Fatalf("half an array rewritten:\n%s", got)
	}
	got = applyEdits(src, cl.formatEdits(3, "textDocument/rangeFormatting", map[string]any{"textDocument": doc, "range": map[string]any{"start": line(2), "end": line(5)}}))
	if !strings.Contains(got, "$a = [\n    1,\n];") {
		t.Fatalf("whole array:\n%s", got)
	}
}

func TestRangeFormattingIndentsFirstLine(t *testing.T) {
	dir := t.TempDir()
	writeTree(t, dir, map[string]string{"composer.json": "{}"})
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + dir, "capabilities": map[string]any{}})
	line := func(n int) map[string]any { return map[string]any{"line": n, "character": 0} }
	for i, tc := range []struct {
		name, src, want string
		first, last     int
	}{
		{"indentation", "<?php\n\nfunction f($x)\n{\n      return $x;\n}\n", "<?php\n\nfunction f($x)\n{\n    return $x;\n}\n", 4, 5},
		{"blank line above", "<?php\n\nfunction f($x)\n{\n    $x++;\n\n\n  return $x;\n}\n", "<?php\n\nfunction f($x)\n{\n    $x++;\n\n\n    return $x;\n}\n", 7, 8},
		{"comment line", "<?php\n\nfunction f($x)\n{\n  // c\n    return $x;\n}\n", "<?php\n\nfunction f($x)\n{\n    // c\n    return $x;\n}\n", 4, 5},
		{"from a blank line", "<?php\n\nfunction f($x)\n{\n    $x++;\n\n\n  return $x;\n}\n", "<?php\n\nfunction f($x)\n{\n    $x++;\n\n\n    return $x;\n}\n", 5, 8},
		{"inside a docblock", "<?php\n\nfunction f($x)\n{\n      /**\n       * Doc.\n       */\n    return $x;\n}\n", "<?php\n\nfunction f($x)\n{\n      /**\n       * Doc.\n       */\n    return $x;\n}\n", 5, 6},
		{"already indented", "<?php\n\nfunction f($x)\n{  \n    return $x;\n}\n", "<?php\n\nfunction f($x)\n{  \n    return $x;\n}\n", 4, 5},
		{"joined to the line above", "<?php\n\nif ($a) {\n}\nelse {\n}\n", "<?php\n\nif ($a) {\n}\nelse {\n}\n", 4, 5},
	} {
		uri := "file://" + filepath.Join(dir, tc.name+".php")
		cl.open(uri, tc.src)
		params := map[string]any{"textDocument": map[string]any{"uri": uri}, "range": map[string]any{"start": line(tc.first), "end": line(tc.last)}}
		if got := applyEdits(tc.src, cl.formatEdits(10+i, "textDocument/rangeFormatting", params)); got != tc.want {
			t.Errorf("%s: got\n%s\nwant\n%s", tc.name, got, tc.want)
		}
	}
}
