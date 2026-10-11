package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/inspection/rules/nestednotoperators"
	"custos/internal/php/syntax"
	"custos/internal/project"
)

func fixture(t testing.TB) (*Server, context.Context, string) {
	t.Helper()
	root := t.TempDir()
	storage := t.TempDir()
	if err := os.Chmod(storage, 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "a.php"), "<?php // é\n$x = !!!$token; function sample(): int { return 1; }\n")
	writeFile(t, filepath.Join(root, "custos.json"), `{"php":"8.2"}`)
	s, err := New(Options{Default: "default", Storage: storage, Workspaces: []Workspace{{ID: "default", Root: root, Readers: []string{"local", "alice", "bob"}, Writers: []string{"local", "alice"}}}}, []analysis.Rule{nestednotoperators.New()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(s.Close)
	return s, context.WithValue(context.WithValue(context.Background(), identityKey{}, "local"), writeScopeKey{}, true), root
}

func writeFile(t testing.TB, path, data string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
}

func invoke(ctx context.Context, t testing.TB, s *Server, name string, in Input) Output {
	t.Helper()
	out, err := s.call(ctx, name, in)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return out
}

func rejected(ctx context.Context, t *testing.T, s *Server, name string, in Input) {
	t.Helper()
	if _, err := s.call(ctx, name, in); err == nil {
		t.Fatalf("%s unexpectedly allowed: %+v", name, in)
	}
}

func TestWorkflow(t *testing.T) {
	s, ctx, root := fixture(t)
	invoke(ctx, t, s, "workspace", Input{})
	invoke(ctx, t, s, "rules", Input{Rule: "NestedNotOperators"})
	a := invoke(ctx, t, s, "analyze", Input{})
	cp := a["checkpointId"].(string)
	findings := invoke(ctx, t, s, "findings", Input{CheckpointID: cp})
	if findings["total"].(int) != 1 {
		t.Fatal(findings)
	}
	syms := invoke(ctx, t, s, "symbols", Input{CheckpointID: cp, Query: "sample"})
	if syms["total"].(int) != 1 {
		t.Fatal(syms)
	}
	src, err := os.ReadFile(filepath.Join(root, "a.php"))
	if err != nil {
		t.Fatal(err)
	}
	types := invoke(ctx, t, s, "type", Input{CheckpointID: cp, Path: "a.php", Offset: uint32(strings.Index(string(src), "1;"))})
	if types["phpDocAware"] != "int" {
		t.Fatal(types)
	}
	p := invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: cp, Choices: []Choice{{FindingID: "0"}}})
	id := p["planId"].(string)
	invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: cp, PlanID: id})
	bob := context.WithValue(ctx, identityKey{}, "bob")
	rejected(bob, t, s, "findings", Input{CheckpointID: cp})
	rejected(bob, t, s, "apply_fixes", Input{PlanID: id})
	alice := context.WithValue(context.WithValue(ctx, identityKey{}, "alice"), writeScopeKey{}, false)
	rejected(alice, t, s, "apply_fixes", Input{PlanID: id})
	applied := invoke(ctx, t, s, "apply_fixes", Input{PlanID: id})
	if applied["partialFailure"] != false {
		t.Fatal(applied)
	}
	src, err = os.ReadFile(filepath.Join(root, "a.php"))
	if err != nil || strings.Contains(string(src), "!!!") {
		t.Fatal(string(src), err)
	}
	rejected(ctx, t, s, "apply_fixes", Input{PlanID: id})
	rejected(ctx, t, s, "findings", Input{CheckpointID: cp})
	invoke(ctx, t, s, "checkpoints", Input{})
	invoke(ctx, t, s, "checkpoints", Input{CheckpointID: cp, Action: "resume"})
	invoke(ctx, t, s, "checkpoints", Input{CheckpointID: cp, Action: "delete"})
}

func TestRecoveryAndInputs(t *testing.T) {
	s, ctx, root := fixture(t)
	if err := os.Mkdir(filepath.Join(root, "vendor"), 0o700); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(root, "vendor", "v.php"), "<?php function vendor_decl(): int {return 2;}")
	a := invoke(ctx, t, s, "analyze", Input{Paths: []string{"a.php"}})
	cp := a["checkpointId"].(string)
	options := s.options
	s.Close()
	recovered, err := New(options, s.rules)
	if err != nil {
		t.Fatal(err)
	}
	defer recovered.Close()
	invoke(ctx, t, recovered, "findings", Input{CheckpointID: cp})
	invoke(ctx, t, recovered, "checkpoints", Input{CheckpointID: cp, Action: "resume"})
	writeFile(t, filepath.Join(root, "vendor", "v.php"), "<?php function vendor_decl(): string {return '2';}")
	rejected(ctx, t, recovered, "prepare_fixes", Input{CheckpointID: cp, Iterative: true})
	if !recovered.checkpoints[cp].Stale {
		t.Fatal("vendor change not marked stale")
	}
	rejected(ctx, t, recovered, "analyze", Input{Paths: []string{"../outside.php"}})
	if err := os.Symlink(t.TempDir(), filepath.Join(root, "link")); err != nil {
		t.Fatal(err)
	}
	rejected(ctx, t, recovered, "analyze", Input{Paths: []string{"link"}})
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	rejected(cancelled, t, recovered, "analyze", Input{})
}

func TestBufferAndSelections(t *testing.T) {
	s, ctx, _ := fixture(t)
	buffer := "<?php $x=!!!$a;"
	a := invoke(ctx, t, s, "analyze", Input{Buffer: &buffer, Path: "unsaved.php"})
	cp := a["checkpointId"].(string)
	p := invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: cp, Iterative: true})
	if p["previewOnly"] != true {
		t.Fatal(p)
	}
	rejected(ctx, t, s, "apply_fixes", Input{PlanID: p["planId"].(string)})
	rejected(ctx, t, s, "analyze", Input{Buffer: &buffer, Path: "../bad.php"})
	rejected(ctx, t, s, "prepare_fixes", Input{CheckpointID: cp})
	rejected(ctx, t, s, "prepare_fixes", Input{CheckpointID: cp, Choices: []Choice{{FindingID: "no"}}})
	rejected(ctx, t, s, "prepare_fixes", Input{CheckpointID: cp, Choices: []Choice{{FindingID: "0", Fix: 99}}})
	rejected(ctx, t, s, "type", Input{CheckpointID: cp, Path: "unsaved.php", Offset: 1000})
	rejected(ctx, t, s, "checkpoints", Input{Action: "invalid"})
	rejected(ctx, t, s, "findings", Input{WorkspaceID: "unknown", CheckpointID: cp})
	s.options.Default = ""
	rejected(ctx, t, s, "analyze", Input{})
	invoke(ctx, t, s, "analyze", Input{WorkspaceID: "default"})
}

func TestPages(t *testing.T) {
	ctx := context.WithValue(context.Background(), identityKey{}, "local")
	items := make([]any, 250)
	for i := range items {
		items[i] = fmt.Sprintf("item %d", i)
	}
	first, err := page(ctx, "one", Input{}, items)
	if err != nil || len(first["items"].([]any)) != 20 {
		t.Fatal(first, err)
	}
	cursor := first["nextCursor"].(string)
	second, err := page(ctx, "one", Input{Cursor: cursor}, items)
	if err != nil || second["items"].([]any)[0] != "item 20" {
		t.Fatal(second, err)
	}
	for _, in := range []Input{{Limit: -1}, {Limit: 101}, {Cursor: "???"}, {Cursor: cursor, Query: "changed"}} {
		if _, err := page(ctx, "one", in, items); err == nil {
			t.Fatal(in)
		}
	}
	if _, err := page(context.WithValue(ctx, identityKey{}, "other"), "one", Input{Cursor: cursor}, items); err == nil {
		t.Fatal("cursor identity leaked")
	}
	large := make([]any, 100)
	for i := range large {
		large[i] = strings.Repeat("a", 2048)
	}
	out, err := page(ctx, "big", Input{Limit: 100}, large)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := json.Marshal(out)
	if len(b) > resultLimit || out["nextCursor"] == nil {
		t.Fatal(len(b))
	}
	if _, err := page(ctx, "big", Input{}, []any{strings.Repeat("x", resultLimit)}); err == nil {
		t.Fatal("oversized item allowed")
	}
}

func TestProtocolRevisions(t *testing.T) {
	for _, revision := range sdk.SupportedProtocolVersions() {
		t.Run(revision, func(t *testing.T) {
			s, ctx, _ := fixture(t)
			serverTransport, clientTransport := sdk.NewInMemoryTransports()
			ss, err := s.Protocol("test").Connect(ctx, serverTransport, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer ss.Close()
			client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "1"}, nil)
			session, err := client.Connect(ctx, clientTransport, &sdk.ClientSessionOptions{ProtocolVersion: revision})
			if err != nil {
				t.Fatal(err)
			}
			defer session.Close()
			list, err := session.ListTools(ctx, nil)
			if err != nil || len(list.Tools) != 9 {
				t.Fatal(list, err)
			}
			result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: "custos_analyze", Arguments: map[string]any{}})
			if err != nil || result.IsError {
				t.Fatal(result, err)
			}
			var text map[string]any
			if err := json.Unmarshal([]byte(result.Content[0].(*sdk.TextContent).Text), &text); err != nil || text["checkpointId"] == nil {
				t.Fatal(text, err)
			}
		})
	}
}

func TestHTTPPolicy(t *testing.T) {
	s, _, _ := fixture(t)
	token := strings.Repeat("a", 48)
	s.options.Credentials = []Credential{{SHA256: project.Fingerprint([]byte(token)), Identity: "alice", Expires: time.Now().Add(time.Hour), Scopes: []string{"custos:read"}}}
	s.options.Hosts = []string{"example.test"}
	s.options.Origins = []string{"https://trusted.test"}
	handler, err := s.HTTPHandler("test")
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		host, origin, auth, path string
		status                   int
	}{{"bad.test", "", "Bearer " + token, "/mcp", 403}, {"example.test", "https://evil.test", "Bearer " + token, "/mcp", 403}, {"example.test", "", "", "/mcp", 401}, {"example.test", "", "Bearer wrong", "/mcp", 401}, {"example.test", "", "Bearer " + token, "/unknown", 404}, {"example.test", "", "", "/.well-known/oauth-protected-resource", 404}}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodPost, "http://"+tc.host+tc.path, strings.NewReader(`{}`))
		req.Header.Set("Authorization", tc.auth)
		req.Header.Set("Origin", tc.origin)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != tc.status {
			t.Fatal(tc, rr.Code, rr.Body.String())
		}
	}
	server := httptest.NewServer(handler)
	s.options.Hosts = append(s.options.Hosts, strings.TrimPrefix(server.URL, "http://"))
	defer server.Close()
	client := sdk.NewClient(&sdk.Implementation{Name: "test", Version: "1"}, nil)
	transport := &sdk.StreamableClientTransport{Endpoint: server.URL + "/mcp", HTTPClient: &http.Client{Transport: bearerTransport{token: token}}}
	ctx := context.Background()
	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	result, err := session.CallTool(ctx, &sdk.CallToolParams{Name: "custos_analyze", Arguments: map[string]any{}})
	if err != nil || result.IsError {
		t.Fatal(result, err)
	}
	var data map[string]any
	_ = json.Unmarshal([]byte(result.Content[0].(*sdk.TextContent).Text), &data)
	result, err = session.CallTool(ctx, &sdk.CallToolParams{Name: "custos_prepare_fixes", Arguments: map[string]any{"checkpointId": data["checkpointId"], "iterative": true}})
	if err != nil || result.IsError {
		t.Fatal(result, err)
	}
	_ = json.Unmarshal([]byte(result.Content[0].(*sdk.TextContent).Text), &data)
	result, err = session.CallTool(ctx, &sdk.CallToolParams{Name: "custos_apply_fixes", Arguments: map[string]any{"planId": data["planId"]}})
	if err != nil || !result.IsError {
		t.Fatal("read-only scope wrote", result, err)
	}
}

type bearerTransport struct{ token string }

func (b bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r = r.Clone(r.Context())
	r.Header.Set("Authorization", "Bearer "+b.token)
	return http.DefaultTransport.RoundTrip(r)
}

func TestConflictingEditsAndWrites(t *testing.T) {
	for _, edits := range [][]diagnostic.TextEdit{{{Span: syntax.Span{Start: 3, End: 2}}}, {{Span: syntax.Span{End: 99}}}, {{Span: syntax.Span{Start: 1, End: 3}}, {Span: syntax.Span{Start: 2, End: 4}}}, {{Span: syntax.Span{Start: 1, End: 1}}, {Span: syntax.Span{Start: 1, End: 1}}}} {
		if validateEdits([]byte("abcde"), edits) == nil {
			t.Fatal("conflict accepted")
		}
	}
	s, ctx, root := fixture(t)
	a := invoke(ctx, t, s, "analyze", Input{})
	cp := a["checkpointId"].(string)
	p := invoke(ctx, t, s, "prepare_fixes", Input{CheckpointID: cp, Iterative: true})
	id := p["planId"].(string)
	s.plans[id].Expires = time.Now().Add(-time.Second)
	rejected(ctx, t, s, "apply_fixes", Input{PlanID: id})
	writeFile(t, filepath.Join(root, "a.php"), "<?php $x = 0;")
	rejected(ctx, t, s, "prepare_fixes", Input{CheckpointID: cp, Iterative: true})
	c := change{Path: "a.php", Source: []byte("wrong"), Output: []byte("<?php"), Perm: 0o600}
	if replace(s.workspaces["default"], c) == nil {
		t.Fatal("stale replace")
	}
	if err := os.Remove(filepath.Join(root, "a.php")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(t.TempDir(), "outside"), filepath.Join(root, "a.php")); err != nil {
		t.Fatal(err)
	}
	if replace(s.workspaces["default"], c) == nil {
		t.Fatal("symlink replace")
	}
}

func TestConcurrentCalls(t *testing.T) {
	s, ctx, _ := fixture(t)
	a := invoke(ctx, t, s, "analyze", Input{})
	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() { _, _ = s.call(ctx, "symbols", Input{CheckpointID: a["checkpointId"].(string)}) })
	}
	wg.Wait()
}

func TestStdio(t *testing.T) {
	s, _, _ := fixture(t)
	err := s.ServeStdio(context.Background(), "test", strings.NewReader(""), io.Discard)
	if err != nil {
		t.Fatal(err)
	}
}

func TestCancelledValidationKeepsCheckpoint(t *testing.T) {
	s, ctx, root := fixture(t)
	writeFile(t, filepath.Join(root, "custos.json"), `{"php":"8.2","paths":["./"],"baseline":"./baseline.json"}`)
	writeFile(t, filepath.Join(root, "baseline.json"), `{"version":1,"entries":[]}`)
	id := invoke(ctx, t, s, "analyze", Input{})["checkpointId"].(string)
	cp := s.checkpoints[id]
	if len(cp.Findings) != 1 {
		t.Fatal(cp.Findings)
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err := s.validate(cancelled, s.workspaces["default"], cp); err == nil || cp.Stale {
		t.Fatal("cancellation invalidated the checkpoint", err)
	}
	if out := invoke(ctx, t, s, "findings", Input{CheckpointID: id}); out["total"] != 1 {
		t.Fatal(out)
	}
}

func TestStartupExpiresForeignStorageFiles(t *testing.T) {
	s, ctx, _ := fixture(t)
	current := invoke(ctx, t, s, "analyze", Input{})["checkpointId"].(string)
	old := time.Now().Add(-8 * 24 * time.Hour)
	foreign, leftover := randomID()+".json", randomID()+"-"+randomID()+".tmp"
	names := []string{foreign, leftover, "notes.json", "short-" + randomID() + ".tmp", strings.Repeat("z", 48) + "-" + randomID() + ".tmp"}
	for _, name := range append(names, "recent-"+foreign) {
		if err := s.storage.WriteFile(name, []byte(`{"root":"/elsewhere"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := s.storage.Chtimes(name, old, old); err != nil {
			t.Fatal(err)
		}
	}
	recent := randomID() + ".json"
	if err := s.storage.WriteFile(recent, []byte(`{"root":"/elsewhere"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := s.load(); err != nil {
		t.Fatal(err)
	}
	for name, kept := range map[string]bool{foreign: false, leftover: false, "notes.json": true, names[3]: true, names[4]: true, recent: true, current + ".json": true} {
		if _, err := s.storage.Lstat(name); (err == nil) != kept {
			t.Fatal(name, kept, err)
		}
	}
}
