package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"custos/internal/analysis"
	"custos/internal/config"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

func frame(body string) string {
	return fmt.Sprintf("Content-Length: %d\r\n\r\n%s", len(body), body)
}

// useRules makes startServer serve rs.
func useRules(t *testing.T, rs ...analysis.Rule) {
	t.Helper()
	registryOverride = func() []analysis.Rule { return rs }
	registry = registryOverride
	t.Cleanup(func() { registryOverride = nil })
}

// waitFor returns the next message matching pred.
func (cl *client) waitFor(what string, pred func(*message) bool) *message {
	cl.t.Helper()
	timeout := time.After(5 * time.Second)
	for {
		select {
		case m, ok := <-cl.out:
			if !ok {
				cl.t.Fatalf("connection closed waiting for %s", what)
			}
			if pred(m) {
				return m
			}
		case <-timeout:
			cl.t.Fatalf("timeout waiting for %s", what)
		}
	}
}

// diagsN waits for diagnostics of uri with n entries.
func (cl *client) diagsN(uri string, n int) {
	cl.t.Helper()
	cl.waitFor(fmt.Sprintf("%d diagnostics of %s", n, uri), func(m *message) bool {
		var p publishDiagnosticsParams
		_ = json.Unmarshal(m.Params, &p)
		return m.Method == "textDocument/publishDiagnostics" && p.URI == uri && len(p.Diagnostics) == n
	})
}

// diagsFor waits for diagnostics of uri and returns them.
func (cl *client) diagsFor(uri string) publishDiagnosticsParams {
	cl.t.Helper()
	var p publishDiagnosticsParams
	cl.waitFor("diagnostics of "+uri, func(m *message) bool {
		if m.Method != "textDocument/publishDiagnostics" {
			return false
		}
		p = publishDiagnosticsParams{}
		_ = json.Unmarshal(m.Params, &p)
		return p.URI == uri
	})
	return p
}

func (cl *client) logContaining(s string) {
	cl.t.Helper()
	cl.waitFor("log "+s, func(m *message) bool {
		return m.Method == "window/logMessage" && strings.Contains(string(m.Params), s)
	})
}

func (cl *client) open(uri, text string) {
	_ = cl.c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": uri, "languageId": "php", "version": 1, "text": text}})
}

func (cl *client) change(uri string, version int, text string) {
	_ = cl.c.notify("textDocument/didChange", map[string]any{"textDocument": map[string]any{"uri": uri, "version": version}, "contentChanges": []any{map[string]any{"text": text}}})
}

func (cl *client) initialize(params map[string]any) {
	cl.t.Helper()
	if _, ok := params["capabilities"]; !ok {
		params["capabilities"] = map[string]any{}
	}
	if res := cl.call(1, "initialize", params); res.Error != nil {
		cl.t.Fatal(res.Error)
	}
}

var zeroRange = map[string]any{"start": map[string]any{"line": 0, "character": 0}, "end": map[string]any{"line": 0, "character": 0}}

// noFixRule reports `;` without a fix.
type noFixRule struct{}

func (noFixRule) ID() string               { return "NestedNotOperators" }
func (noFixRule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNop} }
func (noFixRule) Check(ctx *analysis.Context, n syntax.Node) {
	ctx.Report(n.Span(), "no fix")
}

// emptyFixRule offers a fix without edits (nothing to apply).
type emptyFixRule struct{}

func (emptyFixRule) ID() string               { return "UnnecessarySemicolon" }
func (emptyFixRule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNop} }
func (emptyFixRule) Check(ctx *analysis.Context, n syntax.Node) {
	ctx.Report(n.Span(), "empty fix", analysis.Fix{Title: "Nothing", Edits: func() []analysis.TextEdit { return nil }})
}

// panicFixRule offers a fix whose edits crash.
type panicFixRule struct{}

func (panicFixRule) ID() string               { return "UnnecessarySemicolon" }
func (panicFixRule) Kinds() []syntax.NodeKind { return []syntax.NodeKind{syntax.KNop} }
func (panicFixRule) Check(ctx *analysis.Context, n syntax.Node) {
	ctx.Report(n.Span(), "crashing fix", analysis.Fix{Title: "Crash", Edits: func() []analysis.TextEdit { panic("boom") }})
}

func TestRPCErrorString(t *testing.T) {
	if got := (&rpcError{Code: -1, Message: "x"}).Error(); got != "jsonrpc -1: x" {
		t.Fatal(got)
	}
}

// TestServeEndings: exit, client responses, a failing output and a
// cancelled context end (or not) the loop as documented.
func TestServeEndings(t *testing.T) {
	registry = func() []analysis.Rule { return []analysis.Rule{testRule{}} }
	initReq := frame(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"capabilities":{}}}`)
	var out strings.Builder
	in := frame(`{"jsonrpc":"2.0","id":5,"result":null}`) + frame(`{"jsonrpc":"2.0","method":"exit"}`) + initReq
	if err := Serve(context.Background(), strings.NewReader(in), &out); err != nil || out.Len() != 0 {
		t.Fatalf("exit: %v, output %q", err, out.String())
	}
	// the client stopped reading: replying fails and ends the server
	pr, pw := io.Pipe()
	pr.Close()
	if err := Serve(context.Background(), strings.NewReader(initReq), pw); err == nil {
		t.Fatal("reply to a closed output succeeded")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := Serve(ctx, strings.NewReader(frame(`{"jsonrpc":"2.0","method":"$/setTrace"}`)+initReq), io.Discard); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled: %v", err)
	}
}

func TestDocumentLifecycle(t *testing.T) {
	cl := startServer(t)
	uri := "file:///tmp/life.php"
	// opened before initialize: kept, analysed once an engine exists
	cl.open(uri, "<?php\n;\n")
	_ = cl.c.notify("initialized", map[string]any{}) // before initialize: no engine, no index
	cl.initialize(map[string]any{"workspaceFolders": []any{map[string]any{"uri": "file://" + t.TempDir()}}})
	_ = cl.c.notify("workspace/didChangeConfiguration", map[string]any{"settings": nil})
	if d := cl.diagsFor(uri); len(d.Diagnostics) != 1 {
		t.Fatalf("after configuration: %+v", d)
	}
	for _, n := range []string{"$/cancelRequest", "$/setTrace", "custom/notification"} {
		_ = cl.c.notify(n, map[string]any{})
	}
	_ = cl.c.notify("textDocument/didSave", map[string]any{"textDocument": map[string]any{"uri": uri}}) // no index: nothing to do
	cl.change("file:///tmp/unknown.php", 2, "<?php\n;\n")                                               // not open: ignored
	_ = cl.c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/notes.md", "languageId": "markdown", "version": 1, "text": ";"}})
	res := cl.call(2, "textDocument/codeAction", map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/notes.md"}, "range": zeroRange, "context": map[string]any{}})
	if string(res.Result) != "[]" {
		t.Fatalf("non-PHP document was opened: %s", res.Result)
	}

	// edits within the debounce delay replace the pending analysis; closing
	// cancels it and clears the diagnostics
	Debounce = time.Hour
	cl.change(uri, 2, "<?php\n;;\n")
	cl.change(uri, 3, "<?php\n;;;\n")
	_ = cl.c.notify("textDocument/didClose", map[string]any{"textDocument": map[string]any{"uri": uri}})
	if d := cl.diagsFor(uri); len(d.Diagnostics) != 0 || d.Version != nil {
		t.Fatalf("after close: %+v", d)
	}
}

func TestExecuteCommand(t *testing.T) {
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + t.TempDir()})
	uri := "file:///tmp/cmd.php"
	cl.open(uri, "<?php\n$a = 1;;\n")
	cl.diagsFor(uri)
	for i, tc := range []struct {
		params any
		code   int
	}{
		{"not an object", codeInvalidParams},
		{map[string]any{"command": cmdFixFile}, codeInvalidParams},
		{map[string]any{"command": cmdFixFile, "arguments": []any{42}}, codeInvalidParams},
		{map[string]any{"command": "custos.nope", "arguments": []any{uri}}, codeMethodNotFound},
		{map[string]any{"command": cmdFixRule, "arguments": []any{uri, "NoSuchRule"}}, codeInvalidParams},
	} {
		if res := cl.call(10+i, "workspace/executeCommand", tc.params); res.Error == nil || res.Error.Code != tc.code {
			t.Errorf("case %d: %+v", i, res.Error)
		}
	}
	for i, args := range [][]any{{uri}, {uri, "UnnecessarySemicolon"}} {
		cmd := cmdFixFile
		if len(args) == 2 {
			cmd = cmdFixRule
		}
		raw := json.RawMessage(mustString(30 + i))
		_ = cl.c.write(&message{ID: &raw, Method: "workspace/executeCommand", Params: mustJSON(map[string]any{"command": cmd, "arguments": args})})
		m := cl.waitFor("applyEdit", func(m *message) bool { return m.Method == "workspace/applyEdit" })
		if !strings.Contains(string(m.Params), `?php\n$a = 1;\n"`) {
			t.Fatalf("%s edit: %s", cmd, m.Params)
		}
	}
	// nothing to fix: no edit requested
	clean := "file:///tmp/clean.php"
	cl.open(clean, "<?php\n$a = 1;\n")
	cl.diagsFor(clean)
	if res := cl.call(40, "workspace/executeCommand", map[string]any{"command": cmdFixFile, "arguments": []any{clean}}); res.Error != nil || string(res.Result) != "null" {
		t.Fatalf("clean file: %s %+v", res.Result, res.Error)
	}
}

// TestCodeActionVariants: findings without fixes give no action; without
// resolve support the fix-all edit is inlined; fixes with nothing to apply
// give no edit; requests with malformed params are rejected.
func TestCodeActionVariants(t *testing.T) {
	useRules(t, testRule{}, noFixRule{})
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + t.TempDir()})
	uri := "file:///tmp/ca.php"
	cl.open(uri, "<?php\n;\n")
	if d := cl.diagsFor(uri); len(d.Diagnostics) != 2 {
		t.Fatalf("diagnostics: %+v", d)
	}
	res := cl.call(2, "textDocument/codeAction", map[string]any{"textDocument": map[string]any{"uri": uri}, "range": zeroRange, "context": map[string]any{"only": []string{"source"}}})
	var actions []codeAction
	_ = json.Unmarshal(res.Result, &actions)
	if len(actions) != 1 || actions[0].Kind != kindFixAll || actions[0].Edit == nil || actions[0].Edit.DocumentChanges[0].Edits[0].NewText != "<?php\n\n" {
		t.Fatalf("fix-all: %s", res.Result)
	}
	for i, m := range []string{"codeAction/resolve", "workspace/executeCommand", "textDocument/codeAction"} {
		if res := cl.call(3+i, m, "not an object"); res.Error == nil || res.Error.Code != codeInvalidParams {
			t.Errorf("%s with bad params: %+v", m, res.Error)
		}
	}
}

func TestFixAllResolve(t *testing.T) {
	resolveCaps := map[string]any{"textDocument": map[string]any{"codeAction": map[string]any{"resolveSupport": map[string]any{"properties": []string{"edit"}}}}}
	fixAll := func(cl *client, uri string) *message {
		res := cl.call(2, "textDocument/codeAction", map[string]any{"textDocument": map[string]any{"uri": uri}, "range": zeroRange, "context": map[string]any{"only": []string{kindFixAll}}})
		var actions []codeAction
		_ = json.Unmarshal(res.Result, &actions)
		if len(actions) != 1 || actions[0].Edit != nil {
			t.Fatalf("fix-all action: %s", res.Result)
		}
		return cl.call(3, "codeAction/resolve", actions[0])
	}
	t.Run("applies", func(t *testing.T) {
		cl := startServer(t)
		cl.initialize(map[string]any{"rootUri": "file://" + t.TempDir(), "capabilities": resolveCaps})
		uri := "file:///tmp/all.php"
		cl.open(uri, "<?php\n;;\n")
		cl.diagsFor(uri)
		res := fixAll(cl, uri)
		var a codeAction
		_ = json.Unmarshal(res.Result, &a)
		if res.Error != nil || a.Edit == nil || a.Edit.DocumentChanges[0].Edits[0].NewText != "<?php\n\n" {
			t.Fatalf("resolved: %s %+v", res.Result, res.Error)
		}
	})
	t.Run("nothing to apply", func(t *testing.T) {
		useRules(t, emptyFixRule{})
		cl := startServer(t)
		cl.initialize(map[string]any{"rootUri": "file://" + t.TempDir(), "capabilities": resolveCaps})
		uri := "file:///tmp/empty.php"
		cl.open(uri, "<?php\n;\n")
		cl.diagsFor(uri)
		if res := fixAll(cl, uri); res.Error == nil || res.Error.Code != codeRequestFailed {
			t.Fatalf("empty fix-all: %s %+v", res.Result, res.Error)
		}
	})
}

// TestPanicIsContained: a fix crashing while a request is handled gives an
// internal error reply and a log message; the server keeps serving.
func TestPanicIsContained(t *testing.T) {
	useRules(t, panicFixRule{})
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + t.TempDir()})
	uri := "file:///tmp/panic.php"
	cl.open(uri, "<?php\n;\n")
	cl.diagsFor(uri)
	raw := json.RawMessage("2")
	_ = cl.c.write(&message{ID: &raw, Method: "textDocument/codeAction", Params: mustJSON(map[string]any{"textDocument": map[string]any{"uri": uri}, "range": zeroRange, "context": map[string]any{}})})
	cl.logContaining("panicked: boom")
	res := cl.waitFor("reply", func(m *message) bool { return m.ID != nil && string(*m.ID) == "2" })
	if res.Error == nil || res.Error.Code != codeInternalError {
		t.Fatalf("crashing fix: %s %+v", res.Result, res.Error)
	}
	if res := cl.call(3, "shutdown", nil); res.Error != nil {
		t.Fatal(res.Error)
	}
}

func TestInitializeVariants(t *testing.T) {
	t.Run("rootPath", func(t *testing.T) {
		cl := startServer(t)
		cl.initialize(map[string]any{"rootPath": t.TempDir()})
	})
	fails := func(t *testing.T, params map[string]any, want string) {
		t.Helper()
		cl := startServer(t)
		params["capabilities"] = map[string]any{}
		res := cl.call(1, "initialize", params)
		if res.Error == nil || res.Error.Code != codeRequestFailed || !strings.Contains(res.Error.Message, want) {
			t.Fatalf("got %+v, want an error containing %q", res.Error, want)
		}
	}
	t.Run("invalid custos.json", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte("{"), 0o644); err != nil {
			t.Fatal(err)
		}
		fails(t, map[string]any{"rootUri": "file://" + dir}, config.FileName)
	})
	t.Run("options not an object", func(t *testing.T) {
		fails(t, map[string]any{"rootUri": "file://" + t.TempDir(), "initializationOptions": 5}, "initializationOptions")
	})
	t.Run("invalid option value", func(t *testing.T) {
		fails(t, map[string]any{"rootUri": "file://" + t.TempDir(), "initializationOptions": map[string]any{"php": "5.2"}}, "unsupported version")
	})
	t.Run("rule missing from the catalogue", func(t *testing.T) {
		useRules(t, fakeIDRule{})
		fails(t, map[string]any{"rootUri": "file://" + t.TempDir()}, "missing from catalogue")
	})
}

type fakeIDRule struct{ testRule }

func (fakeIDRule) ID() string { return "NotInCatalogue" }

// TestSettings: client settings override the severity (diagnostic levels
// follow), invalid settings are logged and the previous engine stays.
func TestSettings(t *testing.T) {
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + t.TempDir(), "initializationOptions": map[string]any{"rules": map[string]any{"UnnecessarySemicolon": map[string]any{"severity": "error"}}}})
	uri := "file:///tmp/sev.php"
	cl.open(uri, "<?php\n;\n")
	if d := cl.diagsFor(uri); len(d.Diagnostics) != 1 || d.Diagnostics[0].Severity != 1 {
		t.Fatalf("error severity: %+v", d)
	}
	_ = cl.c.notify("workspace/didChangeConfiguration", map[string]any{"settings": map[string]any{"rules": map[string]any{"UnnecessarySemicolon": map[string]any{"severity": "warning"}}}})
	if d := cl.diagsFor(uri); len(d.Diagnostics) != 1 || d.Diagnostics[0].Severity != 2 {
		t.Fatalf("warning severity: %+v", d)
	}
	_ = cl.c.notify("workspace/didChangeConfiguration", map[string]any{"settings": map[string]any{"comparisonStyle": "sideways"}})
	cl.logContaining("configuration: config: comparisonStyle")
	if d := cl.diagsFor(uri); len(d.Diagnostics) != 1 || d.Diagnostics[0].Severity != 2 {
		t.Fatalf("after invalid settings: %+v", d)
	}
}

// TestMergeOverrides: client settings replace the PHP target, style, short
// tags and per-rule settings, but what to analyse (paths, exclusions,
// baseline) stays the project's.
func TestMergeOverrides(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(`{"php": "8.1", "comparisonStyle": "yoda", "paths": ["src"], "exclude": ["gen"], "baseline": "bl.json",
		"rules": {"NestedNotOperators": {"enabled": false}, "UnnecessarySemicolon": {"severity": "info"}}}`), 0o644); err != nil {
		t.Fatal(err)
	}
	base, err := config.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	m, err := mergeOverrides(base, config.File{Rules: map[string]config.RuleSettings{"UnnecessarySemicolon": {Severity: "error"}}})
	if err != nil {
		t.Fatal(err)
	}
	if m.PHP != phpver.PHP81 || m.ComparisonStyle != analysis.StyleYoda || m.ShortOpenTag ||
		strings.Join(m.Paths, ",") != "src" || !strings.HasSuffix(strings.Join(m.Exclude, ","), ",gen") || m.Baseline != "bl.json" ||
		m.Rules["UnnecessarySemicolon"].Severity != "error" || m.Rules["NestedNotOperators"].Enabled == nil {
		t.Fatalf("merged: %+v", m)
	}
	yes := true
	m, err = mergeOverrides(base, config.File{PHP: "7.4", ComparisonStyle: "regular", ShortOpenTag: &yes})
	if err != nil || m.PHP != phpver.PHP74 || m.ComparisonStyle != analysis.StyleRegular || !m.ShortOpenTag {
		t.Fatalf("overrides: %+v %v", m, err)
	}
}

func TestProjectIndex(t *testing.T) {
	useRules(t, unknownFuncRule{})
	dir := t.TempDir()
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + dir})
	_ = cl.c.notify("initialized", map[string]any{})
	cl.logContaining("indexed")

	user := "file://" + filepath.Join(dir, "user.php")
	lib := "file://" + filepath.Join(dir, "lib.php")
	cl.open(user, "<?php\nhelper();\n")
	cl.diagsN(user, 1)
	// saving an (unsaved-to-disk) buffer indexes its symbols
	cl.open(lib, "<?php\nfunction helper() {}\n")
	cl.diagsFor(lib)
	_ = cl.c.notify("textDocument/didSave", map[string]any{"textDocument": map[string]any{"uri": lib}})
	_ = cl.c.notify("textDocument/didSave", map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/not-open.php"}})
	// the save alone re-analyses the other open documents
	cl.diagsN(user, 0)
	cl.change(user, 2, "<?php\nhelper();\n")
	cl.diagsN(user, 0)
	// non-PHP changes are ignored; an unreadable changed path leaves the index
	_ = cl.c.notify("workspace/didChangeWatchedFiles", map[string]any{"changes": []any{
		map[string]any{"uri": "file://" + filepath.Join(dir, "README.md"), "type": 2},
		map[string]any{"uri": "file://" + dir + "/gone.php", "type": 2},
	}})
	cl.diagsN(user, 0)
	// a new engine (settings change) keeps the built index
	_ = cl.c.notify("workspace/didChangeConfiguration", map[string]any{"settings": map[string]any{"php": "8.2"}})
	cl.diagsN(user, 0)
}

func TestIndexDiscoveryError(t *testing.T) {
	useRules(t, unknownFuncRule{})
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, config.FileName), []byte(`{"paths": ["missing"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + dir})
	_ = cl.c.notify("initialized", map[string]any{})
	cl.logContaining("index: ")
}

// whiteBox returns a server writing to w, configured on dir with rs.
func whiteBox(t *testing.T, dir string, w io.Writer, rs ...analysis.Rule) *Server {
	t.Helper()
	useRules(t, rs...)
	s := &Server{c: newConn(strings.NewReader(""), w), docs: map[string]*document{}, sem: make(chan struct{}, 2), root: dir}
	if err := s.configure(); err != nil {
		t.Fatal(err)
	}
	return s
}

// TestChangesQueuedWhileIndexing: file changes reported while the initial
// index is built are replayed once it is ready.
func TestChangesQueuedWhileIndexing(t *testing.T) {
	dir := t.TempDir()
	s := whiteBox(t, dir, io.Discard, unknownFuncRule{})
	if !s.beginIndexing() {
		t.Fatal("semantic rule enabled but no index wanted")
	}
	helper := filepath.Join(dir, "helper.php")
	s.watchedFilesChanged([]fileChange{{URI: "file://" + helper, Type: 1}})
	if len(s.pending) != 1 {
		t.Fatalf("pending: %+v", s.pending)
	}
	if err := os.WriteFile(helper, []byte("<?php\nfunction helper() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s.buildIndex()
	if s.indexing || s.pending != nil || s.index.Function("helper", phpver.Default) == nil {
		t.Fatalf("after build: indexing=%v pending=%v", s.indexing, s.pending)
	}
}

// TestApplyEditRequestFails: the client connection failing while asking
// it to apply an edit is reported as an internal error.
func TestApplyEditRequestFails(t *testing.T) {
	pr, pw := io.Pipe()
	pr.Close()
	s := whiteBox(t, t.TempDir(), pw, testRule{})
	uri := "file:///tmp/x.php"
	s.open(didOpenParams{TextDocument: textDocumentItem{URI: uri, LanguageID: "php", Version: 1, Text: "<?php\n;\n"}})
	if _, rerr := s.executeCommand(executeCommandParams{Command: cmdFixFile, Arguments: []json.RawMessage{mustJSON(uri)}}); rerr == nil || rerr.Code != codeInternalError {
		t.Fatalf("got %+v", rerr)
	}
}

// gateRule blocks its first Check until released, to hold an analysis.
type gateRule struct{ testRule }

var gate struct {
	sync.Mutex
	armed            bool
	started, release chan struct{}
}

func (r gateRule) Check(ctx *analysis.Context, n syntax.Node) {
	gate.Lock()
	armed := gate.armed
	gate.armed = false
	gate.Unlock()
	if armed {
		close(gate.started)
		<-gate.release
	}
	r.testRule.Check(ctx, n)
}

// TestStaleAnalysisDropped: an analysis finishing after the document
// changed again publishes nothing.
func TestStaleAnalysisDropped(t *testing.T) {
	useRules(t, gateRule{})
	cl := startServer(t)
	cl.initialize(map[string]any{"rootUri": "file://" + t.TempDir()})
	uri := "file:///tmp/stale.php"
	cl.open(uri, "<?php\n;\n")
	cl.diagsFor(uri)

	gate.Lock()
	gate.armed, gate.started, gate.release = true, make(chan struct{}), make(chan struct{})
	gate.Unlock()
	cl.change(uri, 2, "<?php\n;;\n")
	<-gate.started // version 2 is being analysed
	cl.change(uri, 3, "<?php\n;;;\n")
	cl.call(9, "shutdown", nil) // the change to version 3 was handled
	close(gate.release)
	d := cl.diagsFor(uri)
	if d.Version == nil || *d.Version != 3 || len(d.Diagnostics) != 3 {
		t.Fatalf("published %+v", d)
	}
}
