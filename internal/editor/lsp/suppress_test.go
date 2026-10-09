package lsp

import (
	"encoding/json"
	"strings"
	"testing"
)

// openWith starts a server (resolve support optional), opens src and waits
// for its diagnostics.
func openWith(t *testing.T, resolve bool, src string) (*client, string) {
	t.Helper()
	cl := startServer(t)
	caps := map[string]any{}
	if resolve {
		caps = map[string]any{"textDocument": map[string]any{"codeAction": map[string]any{"resolveSupport": map[string]any{"properties": []string{"edit"}}}}}
	}
	if res := cl.call(1, "initialize", map[string]any{"rootUri": "file://" + t.TempDir(), "capabilities": caps}); res.Error != nil {
		t.Fatal(res.Error)
	}
	_ = cl.c.notify("initialized", map[string]any{})
	uri := "file:///tmp/s.php"
	_ = cl.c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": uri, "languageId": "php", "version": 1, "text": src}})
	cl.wait("textDocument/publishDiagnostics")
	return cl, uri
}

func actionsAt(t *testing.T, cl *client, uri string, line int) []codeAction {
	t.Helper()
	res := cl.call(2, "textDocument/codeAction", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"range":        map[string]any{"start": map[string]any{"line": line, "character": 0}, "end": map[string]any{"line": line, "character": 99}},
		"context":      map[string]any{"diagnostics": []any{}},
	})
	var actions []codeAction
	if err := json.Unmarshal(res.Result, &actions); err != nil {
		t.Fatal(err)
	}
	return actions
}

func suppressAction(actions []codeAction) *codeAction {
	for i := range actions {
		if strings.HasPrefix(actions[i].Title, "Suppress ") {
			return &actions[i]
		}
	}
	return nil
}

const suppressSrc = "<?php\n$a = 1;\nfunction f() {\n    ;\n}\n"

func TestSuppressActionResolved(t *testing.T) {
	cl, uri := openWith(t, true, suppressSrc)
	actions := actionsAt(t, cl, uri, 3)
	a := suppressAction(actions)
	if a == nil || a.Title != "Suppress UnnecessarySemicolon for this statement" || a.Edit != nil || a.IsPreferred {
		t.Fatalf("suppress action: %+v", actions)
	}
	if actions[len(actions)-2].Title != a.Title {
		t.Errorf("suppressions should follow the fixes, before fix-all: %+v", actions)
	}
	res := cl.call(3, "codeAction/resolve", *a)
	var resolved codeAction
	_ = json.Unmarshal(res.Result, &resolved)
	if resolved.Edit == nil {
		t.Fatalf("resolve: %s", res.Result)
	}
	e := resolved.Edit.DocumentChanges[0].Edits[0]
	if e.NewText != "    // @custos-ignore UnnecessarySemicolon\n" || e.Range.Start != (position{3, 0}) || e.Range.End != (position{3, 0}) {
		t.Fatalf("edit: %+v", e)
	}
}

func TestSuppressActionInline(t *testing.T) {
	// A second finding of the same rule elsewhere stays reported.
	cl, uri := openWith(t, false, suppressSrc+"function g() {\n    ;\n}\n")
	a := suppressAction(actionsAt(t, cl, uri, 3))
	if a == nil || a.Edit == nil {
		t.Fatalf("inline edit expected: %+v", a)
	}
}

func TestSuppressActionRefused(t *testing.T) {
	// The file's first statement: a comment there would be file-wide.
	cl, uri := openWith(t, true, "<?php\n;\n")
	if a := suppressAction(actionsAt(t, cl, uri, 1)); a != nil {
		t.Errorf("offered at file level: %+v", a)
	}
	// Two findings share the only annotatable statement: suppressing one
	// would silence both, so it is not offered.
	cl, uri = openWith(t, true, "<?php\n$a = 1;\nif ($a) { ; ; }\n")
	if a := suppressAction(actionsAt(t, cl, uri, 2)); a != nil {
		t.Errorf("offered although it hides two findings: %+v", a)
	}
	// Resolving data that matches no finding, or a finding without a line
	// to annotate, fails cleanly.
	for _, data := range []actionData{
		{URI: uri, Version: 1, Rule: "Other", Start: 1, End: 2, Suppress: true},
		{URI: uri, Version: 1, Rule: "UnnecessarySemicolon", Start: 0, End: 0, Suppress: true},
	} {
		res := cl.call(4, "codeAction/resolve", codeAction{Title: "x", Data: mustJSON(data)})
		if res.Error == nil {
			t.Errorf("resolve %+v: expected an error, got %s", data, res.Result)
		}
	}
	cl, uri = openWith(t, true, "<?php\n;\n")
	res := cl.call(5, "codeAction/resolve", codeAction{Title: "x", Data: mustJSON(actionData{URI: uri, Version: 1, Rule: "UnnecessarySemicolon", Start: 6, End: 7, Suppress: true})})
	if res.Error == nil {
		t.Errorf("file-level suppression resolved: %s", res.Result)
	}
}
