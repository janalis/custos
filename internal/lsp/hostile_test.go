package lsp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"custos/internal/analysis"
)

// startRaw starts a server and returns a client plus the raw stream to it,
// for writing malformed frames.
func startRaw(t *testing.T) (*client, io.Writer) {
	t.Helper()
	registry = func() []analysis.Rule { return []analysis.Rule{testRule{}} }
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
	return cl, cw
}

// nextError waits for an error response.
func (cl *client) nextError() *rpcError {
	cl.t.Helper()
	timeout := time.After(5 * time.Second)
	for {
		select {
		case m, ok := <-cl.out:
			if !ok {
				cl.t.Fatal("connection closed")
			}
			if m.Error != nil {
				return m.Error
			}
		case <-timeout:
			cl.t.Fatal("timeout waiting for an error response")
		}
	}
}

func TestHostileFrames(t *testing.T) {
	old := maxMessageSize
	maxMessageSize = 1 << 10
	defer func() { maxMessageSize = old }()
	cl, raw := startRaw(t)

	// oversized body: rejected and skipped, the stream stays in sync
	big := strings.Repeat("x", 4<<10)
	fmt.Fprintf(raw, "Content-Length: %d\r\n\r\n%s", len(big), big)
	if e := cl.nextError(); e.Code != codeInvalidRequest {
		t.Fatalf("oversized: %+v", e)
	}
	// invalid JSON
	fmt.Fprintf(raw, "Content-Length: 5\r\n\r\n{nope")
	if e := cl.nextError(); e.Code != codeParseError {
		t.Fatalf("invalid json: %+v", e)
	}
	// valid JSON, wrong shape
	body := `{"jsonrpc":"2.0","id":7,"method":"initialize","params":{"rootUri":5}}`
	fmt.Fprintf(raw, "Content-Length: %d\r\n\r\n%s", len(body), body)
	if e := cl.nextError(); e.Code != codeInvalidParams {
		t.Fatalf("wrong types: %+v", e)
	}
	// still serving
	if res := cl.call(1, "initialize", map[string]any{"rootUri": "file://" + t.TempDir(), "capabilities": map[string]any{}}); res.Error != nil {
		t.Fatal(res.Error)
	}
	if res := cl.call(2, "no/such/method", nil); res.Error == nil || res.Error.Code != codeMethodNotFound {
		t.Fatalf("unknown method: %+v", res.Error)
	}
}

func TestHostileRequests(t *testing.T) {
	cl, _ := startRaw(t)
	if res := cl.call(1, "initialize", map[string]any{"rootUri": "file://" + t.TempDir(), "capabilities": map[string]any{}}); res.Error != nil {
		t.Fatal(res.Error)
	}
	uri := "file:///tmp/h.php"
	_ = cl.c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": uri, "languageId": "php", "version": 1, "text": "<?php\n$a = 1;;\n"}})
	cl.wait("textDocument/publishDiagnostics")

	// out-of-range, inverted and negative edit ranges
	for i, r := range []map[string]any{
		{"start": map[string]any{"line": -5, "character": -3}, "end": map[string]any{"line": 0, "character": 2}},
		{"start": map[string]any{"line": 1, "character": 1 << 30}, "end": map[string]any{"line": 1 << 30, "character": 0}},
		{"start": map[string]any{"line": 1, "character": 4}, "end": map[string]any{"line": 0, "character": 1}},
	} {
		_ = cl.c.notify("textDocument/didChange", map[string]any{
			"textDocument":   map[string]any{"uri": uri, "version": 2 + i},
			"contentChanges": []any{map[string]any{"range": r, "text": ";"}},
		})
		cl.wait("textDocument/publishDiagnostics")
	}
	// wrong types in notifications and requests
	_ = cl.c.notify("textDocument/didChange", map[string]any{"textDocument": "x", "contentChanges": 3})
	_ = cl.c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": 1, "text": []int{1}}})
	_ = cl.c.notify("workspace/didChangeWatchedFiles", map[string]any{"changes": "nope"})
	if res := cl.call(2, "textDocument/codeAction", map[string]any{"textDocument": map[string]any{"uri": uri}, "range": "bad"}); res.Error == nil {
		t.Fatalf("codeAction with a bad range: %s", res.Result)
	}
	// unknown documents, bogus action data
	res := cl.call(3, "textDocument/codeAction", map[string]any{
		"textDocument": map[string]any{"uri": "file:///nope.php"},
		"range":        map[string]any{"start": map[string]any{"line": 0, "character": 0}, "end": map[string]any{"line": 0, "character": 0}},
		"context":      map[string]any{"diagnostics": []any{}},
	})
	if res.Error != nil || string(res.Result) != "[]" {
		t.Fatalf("codeAction on unknown doc: %s %+v", res.Result, res.Error)
	}
	for i, data := range []any{
		map[string]any{"uri": "file:///nope.php", "version": 1},
		map[string]any{"uri": uri, "version": 4, "rule": "UnnecessarySemicolon", "start": 0, "end": 1 << 31, "fix": -1},
		"not an object",
	} {
		res = cl.call(10+i, "codeAction/resolve", map[string]any{"title": "x", "data": data})
		if res.Error == nil {
			t.Fatalf("resolve %d: %s", i, res.Result)
		}
	}
	if res = cl.call(20, "workspace/executeCommand", map[string]any{"command": cmdFixFile, "arguments": []any{"file:///nope.php"}}); res.Error == nil {
		t.Fatal("executeCommand on unknown doc")
	}
	if res = cl.call(21, "workspace/executeCommand", map[string]any{"command": cmdFixRule, "arguments": []any{uri, 42}}); res.Error == nil {
		t.Fatal("executeCommand with a bad rule argument")
	}
	if res = cl.call(22, "shutdown", nil); res.Error != nil {
		t.Fatal(res.Error)
	}
}

// TestBadFramingEnds checks that unrecoverable framing errors end the
// server with an error instead of crashing or hanging.
func TestBadFramingEnds(t *testing.T) {
	for name, in := range map[string]string{
		"huge header line":        strings.Repeat("X", 200<<10) + "\r\n\r\n",
		"negative length":         "Content-Length: -4\r\n\r\n{}",
		"non-numeric length":      "Content-Length: lots\r\n\r\n{}",
		"overflowing length":      "Content-Length: 99999999999999999999999\r\n\r\n{}",
		"missing length":          "Foo: bar\r\n\r\n{}",
		"truncated body":          "Content-Length: 100\r\n\r\n{}",
		"huge length, short body": fmt.Sprintf("Content-Length: %d\r\n\r\n{}", 1<<40),
		"max int length":          "Content-Length: 9223372036854775807\r\n\r\n{}",
	} {
		done := make(chan error, 1)
		var out bytes.Buffer
		go func() { done <- Serve(context.Background(), strings.NewReader(in), &out) }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatalf("%s: server did not return", name)
		}
	}
}

func TestOversizedDocument(t *testing.T) {
	cl, _ := startRaw(t)
	if res := cl.call(1, "initialize", map[string]any{"rootUri": "file://" + t.TempDir(), "capabilities": map[string]any{}}); res.Error != nil {
		t.Fatal(res.Error)
	}
	text := "<?php\n" + strings.Repeat(";", 11<<20)
	_ = cl.c.notify("textDocument/didOpen", map[string]any{"textDocument": map[string]any{"uri": "file:///tmp/big.php", "languageId": "php", "version": 1, "text": text}})
	var diags publishDiagnosticsParams
	_ = json.Unmarshal(cl.wait("textDocument/publishDiagnostics").Params, &diags)
	if len(diags.Diagnostics) != 0 {
		t.Fatalf("oversized document: %d diagnostics", len(diags.Diagnostics))
	}
}
