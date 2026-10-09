package lsp

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/catalogue"
	"custos/internal/php/syntax"
)

// testRule flags every `;` empty statement, with a fix deleting it. It uses
// the UnnecessarySemicolon ID so the catalogue lookup succeeds.
type testRule struct{}

func (testRule) ID() string               { return "UnnecessarySemicolon" }
func (testRule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNop} }
func (testRule) Check(ctx *analysis.Context, n syntax.Node) {
	sp := n.Span()
	ctx.Report(sp, "empty statement", diagnostic.Fix{Title: "Remove", Edits: func() []diagnostic.TextEdit {
		return []diagnostic.TextEdit{{Span: sp, NewText: ""}}
	}})
}

type client struct {
	t   *testing.T
	c   *conn
	out chan *message
}

func startServer(t *testing.T) *client {
	t.Helper()
	if registryOverride == nil {
		registry = func() []analysis.Rule { return []analysis.Rule{testRule{}} }
	}
	Debounce = 5 * time.Millisecond
	cr, sw := io.Pipe()
	sr, cw := io.Pipe()
	go func() { _ = Serve(context.Background(), sr, sw); sw.Close() }()
	cl := &client{t: t, c: newConn(cr, cw), out: make(chan *message, 64)}
	go func() {
		for {
			m, err := cl.c.read()
			if err != nil {
				close(cl.out)
				return
			}
			cl.out <- m
		}
	}()
	t.Cleanup(func() { cw.Close() })
	return cl
}

func (cl *client) call(id int, method string, params any) *message {
	cl.t.Helper()
	raw := json.RawMessage(strings.TrimSpace(mustString(id)))
	b, _ := json.Marshal(params)
	if err := cl.c.write(&message{ID: &raw, Method: method, Params: b}); err != nil {
		cl.t.Fatal(err)
	}
	for m := range cl.out {
		if m.ID != nil && string(*m.ID) == string(raw) && m.Method == "" {
			return m
		}
	}
	cl.t.Fatal("connection closed")
	return nil
}

func (cl *client) wait(method string) *message {
	cl.t.Helper()
	timeout := time.After(5 * time.Second)
	for {
		select {
		case m, ok := <-cl.out:
			if !ok {
				cl.t.Fatal("connection closed")
			}
			if m.Method == method {
				return m
			}
		case <-timeout:
			cl.t.Fatalf("timeout waiting for %s", method)
		}
	}
}

func mustString(id int) string { b, _ := json.Marshal(id); return string(b) }

func TestServerRoundTrip(t *testing.T) {
	cl := startServer(t)
	res := cl.call(1, "initialize", map[string]any{
		"rootUri":      "file://" + t.TempDir(),
		"capabilities": map[string]any{"textDocument": map[string]any{"codeAction": map[string]any{"resolveSupport": map[string]any{"properties": []string{"edit"}}}}},
	})
	if res.Error != nil {
		t.Fatal(res.Error)
	}
	_ = cl.c.notify("initialized", map[string]any{})
	uri := "file:///tmp/x.php"
	_ = cl.c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": uri, "languageId": "php", "version": 1, "text": "<?php\n$a = 1;;\n"}})

	var diags publishDiagnosticsParams
	_ = json.Unmarshal(cl.wait("textDocument/publishDiagnostics").Params, &diags)
	if len(diags.Diagnostics) != 1 || diags.Diagnostics[0].Code != "UnnecessarySemicolon" {
		t.Fatalf("diagnostics: %+v", diags)
	}
	d := diags.Diagnostics[0]
	if d.Range.Start != (position{1, 7}) || d.Range.End != (position{1, 8}) {
		t.Fatalf("range: %+v", d.Range)
	}

	// Incremental edit: insert another `;` at the end of line 1.
	_ = cl.c.notify("textDocument/didChange", map[string]any{
		"textDocument":   map[string]any{"uri": uri, "version": 2},
		"contentChanges": []any{map[string]any{"range": map[string]any{"start": map[string]any{"line": 1, "character": 8}, "end": map[string]any{"line": 1, "character": 8}}, "text": ";"}},
	})
	_ = json.Unmarshal(cl.wait("textDocument/publishDiagnostics").Params, &diags)
	if len(diags.Diagnostics) != 2 || diags.Version == nil || *diags.Version != 2 {
		t.Fatalf("after change: %+v", diags)
	}

	res = cl.call(2, "textDocument/codeAction", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"range":        map[string]any{"start": map[string]any{"line": 1, "character": 7}, "end": map[string]any{"line": 1, "character": 7}},
		"context":      map[string]any{"diagnostics": []any{}},
	})
	var actions []codeAction
	_ = json.Unmarshal(res.Result, &actions)
	if len(actions) != 2 || actions[0].Kind != "quickfix" || actions[0].Edit != nil || actions[1].Kind != kindFixAll {
		t.Fatalf("actions: %s", res.Result)
	}
	res = cl.call(3, "codeAction/resolve", actions[0])
	var resolved codeAction
	_ = json.Unmarshal(res.Result, &resolved)
	if resolved.Edit == nil || len(resolved.Edit.DocumentChanges[0].Edits) != 1 {
		t.Fatalf("resolve: %s", res.Result)
	}
	e := resolved.Edit.DocumentChanges[0].Edits[0]
	if e.NewText != "" || e.Range.Start != (position{1, 7}) {
		t.Fatalf("edit: %+v", e)
	}

	res = cl.call(4, "shutdown", nil)
	if res.Error != nil {
		t.Fatal(res.Error)
	}
}

// registryOverride keeps startServer from replacing a test-chosen registry.
var registryOverride func() []analysis.Rule

func TestServerRealRules(t *testing.T) {
	registryOverride, registry = catalogue.All, catalogue.All
	defer func() { registryOverride = nil }()
	cl := startServer(t)
	if res := cl.call(1, "initialize", map[string]any{"rootUri": "file://" + t.TempDir(), "capabilities": map[string]any{}}); res.Error != nil {
		t.Fatal(res.Error)
	}
	uri := "file:///tmp/real.php"
	_ = cl.c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": uri, "languageId": "php", "version": 1, "text": "<?php\n$x = !!$y;\n"}})
	var diags publishDiagnosticsParams
	_ = json.Unmarshal(cl.wait("textDocument/publishDiagnostics").Params, &diags)
	found := false
	for _, d := range diags.Diagnostics {
		if d.Code == "NestedNotOperators" {
			found = true
		}
	}
	if !found {
		t.Fatalf("NestedNotOperators not reported: %+v", diags.Diagnostics)
	}
	// Without resolve support the edit is inlined.
	res := cl.call(2, "textDocument/codeAction", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"range":        map[string]any{"start": map[string]any{"line": 1, "character": 6}, "end": map[string]any{"line": 1, "character": 6}},
		"context":      map[string]any{"diagnostics": []any{}, "only": []string{"quickfix"}},
	})
	var actions []codeAction
	_ = json.Unmarshal(res.Result, &actions)
	if len(actions) == 0 || actions[0].Edit == nil || actions[0].Edit.DocumentChanges[0].Edits[0].NewText != "(bool)$y" {
		t.Fatalf("actions: %s", res.Result)
	}
}

func TestURIToPath(t *testing.T) {
	cases := map[string]string{
		"file:///tmp/a%20b.php": filepath.FromSlash("/tmp/a b.php"),
		"file:///C:/work/x.php": filepath.FromSlash("C:/work/x.php"),
		"untitled:Untitled-1":   "untitled:Untitled-1",
	}
	for in, want := range cases {
		if got := uriToPath(in); got != want {
			t.Errorf("%s: got %s want %s", in, got, want)
		}
	}
}

// unknownFuncRule flags calls to functions missing from the index (test-only,
// semantic so the server builds and maintains the project index).
type unknownFuncRule struct{}

func (unknownFuncRule) ID() string               { return "UnnecessarySemicolon" }
func (unknownFuncRule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KFuncCall} }
func (unknownFuncRule) Semantic()                {}
func (unknownFuncRule) Check(ctx *analysis.Context, n syntax.Node) {
	c := n.(*syntax.FuncCall)
	if nm, ok := c.Name.(*syntax.Name); ok && ctx.Index().Function(nm.Value, ctx.PHP) == nil {
		ctx.ReportNode(c, "unknown function")
	}
}

func TestWatchedFilesUpdateIndex(t *testing.T) {
	registryOverride = func() []analysis.Rule { return []analysis.Rule{unknownFuncRule{}} }
	registry = registryOverride
	defer func() { registryOverride = nil }()
	dir := t.TempDir()
	main := filepath.Join(dir, "main.php")
	if err := os.WriteFile(main, []byte("<?php\nhelper();\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cl := startServer(t)
	res := cl.call(1, "initialize", map[string]any{
		"rootUri":      "file://" + dir,
		"capabilities": map[string]any{"workspace": map[string]any{"didChangeWatchedFiles": map[string]any{"dynamicRegistration": true}}},
	})
	if res.Error != nil {
		t.Fatal(res.Error)
	}
	_ = cl.c.notify("initialized", map[string]any{})
	if reg := cl.wait("client/registerCapability"); !strings.Contains(string(reg.Params), "**/*.php") {
		t.Fatalf("registration: %s", reg.Params)
	}
	uri := "file://" + main
	_ = cl.c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": uri, "languageId": "php", "version": 1, "text": "<?php\nhelper();\n"}})
	waitDiags := func(want int) {
		t.Helper()
		deadline := time.After(5 * time.Second)
		for {
			select {
			case m := <-cl.out:
				if m.Method != "textDocument/publishDiagnostics" {
					continue
				}
				var p publishDiagnosticsParams
				_ = json.Unmarshal(m.Params, &p)
				if p.URI == uri && len(p.Diagnostics) == want {
					return
				}
			case <-deadline:
				t.Fatalf("timeout waiting for %d diagnostics", want)
			}
		}
	}
	waitDiags(1)
	helper := filepath.Join(dir, "helper.php")
	if err := os.WriteFile(helper, []byte("<?php\nfunction helper() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = cl.c.notify("workspace/didChangeWatchedFiles", map[string]any{"changes": []any{map[string]any{"uri": "file://" + helper, "type": 1}}})
	waitDiags(0)
	_ = os.Remove(helper)
	_ = cl.c.notify("workspace/didChangeWatchedFiles", map[string]any{"changes": []any{map[string]any{"uri": "file://" + helper, "type": 3}}})
	waitDiags(1)
}

func TestDiagnosticDataRoundTrip(t *testing.T) {
	cl := startServer(t)
	if res := cl.call(1, "initialize", map[string]any{"rootUri": "file://" + t.TempDir(), "capabilities": map[string]any{}}); res.Error != nil {
		t.Fatal(res.Error)
	}
	uri := "file:///tmp/data.php"
	_ = cl.c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": uri, "languageId": "php", "version": 1, "text": "<?php\n$a = 1;;\n"}})
	var diags publishDiagnosticsParams
	_ = json.Unmarshal(cl.wait("textDocument/publishDiagnostics").Params, &diags)
	if len(diags.Diagnostics) != 1 || diags.Diagnostics[0].Data == nil || diags.Diagnostics[0].Data.Rule != "UnnecessarySemicolon" {
		t.Fatalf("protocolDiagnostic data: %+v", diags.Diagnostics)
	}
	// A request range elsewhere still yields the fix when the protocolDiagnostic (with data) is sent back.
	res := cl.call(2, "textDocument/codeAction", map[string]any{
		"textDocument": map[string]any{"uri": uri},
		"range":        map[string]any{"start": map[string]any{"line": 0, "character": 0}, "end": map[string]any{"line": 0, "character": 0}},
		"context":      map[string]any{"diagnostics": diags.Diagnostics, "only": []string{"quickfix"}},
	})
	var actions []codeAction
	_ = json.Unmarshal(res.Result, &actions)
	if len(actions) != 1 || actions[0].Kind != "quickfix" {
		t.Fatalf("actions via data: %s", res.Result)
	}
}
