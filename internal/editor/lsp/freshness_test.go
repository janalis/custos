package lsp

import (
	"encoding/json"
	"io"
	"sync/atomic"
	"testing"
	"time"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/index"
	"custos/internal/testing/testbudget"
)

type freshnessGate struct {
	testRule
	armed            atomic.Bool
	started, release chan struct{}
}

func (r *freshnessGate) Check(ctx *analysis.Context, n syntax.Node) {
	if r.armed.CompareAndSwap(true, false) {
		close(r.started)
		<-r.release
	}
	r.testRule.Check(ctx, n)
}

func newFreshnessGate() *freshnessGate {
	return &freshnessGate{started: make(chan struct{}), release: make(chan struct{})}
}

func waitFreshness(t *testing.T, done <-chan struct{}) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(testbudget.Of(2 * time.Second)):
		t.Fatal("freshness operation did not complete")
	}
}

func openFresh(s *Server, uri, text string) {
	s.open(didOpenParams{TextDocument: textDocumentItem{URI: uri, LanguageID: "php", Version: 1, Text: text}})
}

func firstFreshAction(t *testing.T, s *Server, uri string) actionData {
	t.Helper()
	actions := s.codeActions(freshActionParams(uri))
	if len(actions) == 0 {
		t.Fatal("missing action")
	}
	var data actionData
	if err := json.Unmarshal(actions[0].Data, &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func TestAnalysisCannotReplaceReopenedDocument(t *testing.T) {
	rule := newFreshnessGate()
	s := whiteBox(t, t.TempDir(), io.Discard, rule)
	uri := "file:///tmp/reopened.php"
	rule.armed.Store(true)
	done := make(chan struct{})
	go func() { defer close(done); openFresh(s, uri, "<?php ;") }()
	waitFreshness(t, rule.started)
	s.close(uri)
	openFresh(s, uri, "<?php ;;")
	s.mu.Lock()
	snapshot := s.docs[uri].snapshot
	s.mu.Unlock()
	close(rule.release)
	waitFreshness(t, done)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.docs[uri].snapshot != snapshot || len(s.docs[uri].findings) != 2 {
		t.Fatal("old document replaced reopened findings")
	}
}

func TestActionsRequireCurrentAnalysisSnapshot(t *testing.T) {
	s := whiteBox(t, t.TempDir(), io.Discard, testRule{})
	uri := "file:///tmp/actions.php"
	openFresh(s, uri, "<?php ;")
	old := firstFreshAction(t, s, uri)
	s.close(uri)
	openFresh(s, uri, "<?php ;")
	if s.buildEdit(old) != nil {
		t.Fatal("reopened document accepted old action")
	}
	current := firstFreshAction(t, s, uri)
	if s.buildEdit(current) == nil {
		t.Fatal("fresh action rejected")
	}
	s.mu.Lock()
	s.engine = s.engine.WithIndex(index.New(nil))
	s.mu.Unlock()
	if len(s.codeActions(freshActionParams(uri))) != 0 || s.buildEdit(current) != nil {
		t.Fatal("outdated engine actions accepted")
	}
	s.analyzeNow(uri)
	if s.buildEdit(current) != nil {
		t.Fatal("old action revived after reanalysis")
	}
	if s.buildEdit(firstFreshAction(t, s, uri)) == nil {
		t.Fatal("new analysis has no edit")
	}
}

func TestLazyFixRevalidatesAfterComputation(t *testing.T) {
	for _, replacement := range []string{"document", "engine", "analysis", "pending-edit"} {
		t.Run(replacement, func(t *testing.T) {
			s := whiteBox(t, t.TempDir(), io.Discard, testRule{})
			uri := "file:///tmp/lazy.php"
			openFresh(s, uri, "<?php ;")
			data := firstFreshAction(t, s, uri)
			started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
			s.mu.Lock()
			original := s.docs[uri].findings[0].Fixes[0].Edits
			s.docs[uri].findings[0].Fixes[0].Edits = func() []diagnostic.TextEdit { close(started); <-release; return original() }
			s.mu.Unlock()
			var edit *workspaceEdit
			go func() { defer close(done); edit = s.buildEdit(data) }()
			waitFreshness(t, started)
			switch replacement {
			case "document":
				s.close(uri)
				openFresh(s, uri, "<?php ;")
			case "engine":
				s.mu.Lock()
				s.engine = s.engine.WithIndex(index.New(nil))
				s.mu.Unlock()
			case "analysis":
				s.analyzeNow(uri)
			case "pending-edit":
				freshChange(s, uri, 1, "<?php ;;")
			}
			close(release)
			waitFreshness(t, done)
			if edit != nil {
				t.Fatal("lazy edit from outdated computation accepted")
			}
		})
	}
}

func TestFixCommandRevalidatesAfterComputation(t *testing.T) {
	rule := newFreshnessGate()
	s := whiteBox(t, t.TempDir(), io.Discard, rule)
	uri := "file:///tmp/command.php"
	openFresh(s, uri, "<?php ;")
	rule.armed.Store(true)
	done := make(chan struct{})
	var failure *rpcError
	go func() {
		defer close(done)
		_, failure = s.executeCommand(executeCommandParams{Command: cmdFixFile, Arguments: []json.RawMessage{mustJSON(uri)}})
	}()
	waitFreshness(t, rule.started)
	s.close(uri)
	openFresh(s, uri, "<?php ;;")
	close(rule.release)
	waitFreshness(t, done)
	if failure == nil || failure.Code != codeRequestFailed {
		t.Fatal(failure)
	}
}

func freshActionParams(uri string) (p codeActionParams) {
	p.TextDocument.URI = uri
	p.Range.End = position{Line: 0, Character: 999}
	return p
}

func TestCurrentSnapshotRejectsUnmatchedFix(t *testing.T) {
	s := whiteBox(t, t.TempDir(), io.Discard, testRule{})
	uri := "file:///tmp/unmatched.php"
	openFresh(s, uri, "<?php ;")
	current := firstFreshAction(t, s, uri)
	for _, fix := range []int{-1, 99} {
		bad := current
		bad.Fix = fix
		if s.buildEdit(bad) != nil {
			t.Fatal("invalid fix accepted")
		}
	}
	bad := current
	bad.Rule = "Other"
	if s.buildEdit(bad) != nil {
		t.Fatal("unmatched finding accepted")
	}
	bad.Suppress = true
	if s.buildEdit(bad) != nil {
		t.Fatal("unmatched suppression accepted")
	}
	s.close(uri)
	openFresh(s, uri, "<?php return 1;")
	s.mu.Lock()
	snapshot := s.docs[uri].snapshot
	s.mu.Unlock()
	if s.buildEdit(actionData{URI: uri, Version: 1, Snapshot: snapshot, All: true}) != nil {
		t.Fatal("empty fix-all offered")
	}
}

func freshChange(s *Server, uri string, version int, text string) {
	var p didChangeParams
	_ = json.Unmarshal(mustJSON(map[string]any{"textDocument": map[string]any{"uri": uri, "version": version}, "contentChanges": []any{map[string]any{"text": text}}}), &p)
	// This white-box edit must remain pending until the test refreshes it.
	delay := Debounce
	Debounce = time.Hour
	s.change(p)
	Debounce = delay
	s.mu.Lock()
	s.docs[uri].timer.Stop()
	s.mu.Unlock()
}

func TestAnalysisRejectsReusedClientVersion(t *testing.T) {
	rule := newFreshnessGate()
	s := whiteBox(t, t.TempDir(), io.Discard, rule)
	uri := "file:///tmp/reused-version.php"
	rule.armed.Store(true)
	done := make(chan struct{})
	go func() { defer close(done); openFresh(s, uri, "<?php ;") }()
	waitFreshness(t, rule.started)
	freshChange(s, uri, 1, "<?php ;;")
	s.analyzeNow(uri)
	s.mu.Lock()
	snapshot := s.docs[uri].snapshot
	s.mu.Unlock()
	close(rule.release)
	waitFreshness(t, done)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.docs[uri].snapshot != snapshot || len(s.docs[uri].findings) != 2 {
		t.Fatal("reused client version accepted old analysis")
	}
}

func TestActionListRevalidatesAfterComputation(t *testing.T) {
	s := whiteBox(t, t.TempDir(), io.Discard, testRule{})
	uri := "file:///tmp/list.php"
	openFresh(s, uri, "<?php ;")
	started, release, done := make(chan struct{}), make(chan struct{}), make(chan struct{})
	s.mu.Lock()
	original := s.docs[uri].findings[0].Fixes[0].Edits
	s.docs[uri].findings[0].Fixes[0].Edits = func() []diagnostic.TextEdit { close(started); <-release; return original() }
	s.mu.Unlock()
	var actions []codeAction
	go func() { defer close(done); actions = s.codeActions(freshActionParams(uri)) }()
	waitFreshness(t, started)
	freshChange(s, uri, 1, "<?php ;;")
	close(release)
	waitFreshness(t, done)
	if len(actions) != 0 {
		t.Fatal("stale action list returned")
	}
}

func TestFixCommandRejectsReusedClientVersion(t *testing.T) {
	rule := newFreshnessGate()
	s := whiteBox(t, t.TempDir(), io.Discard, rule)
	uri := "file:///tmp/command-version.php"
	openFresh(s, uri, "<?php ;")
	rule.armed.Store(true)
	done := make(chan struct{})
	var failure *rpcError
	go func() {
		defer close(done)
		_, failure = s.executeCommand(executeCommandParams{Command: cmdFixFile, Arguments: []json.RawMessage{mustJSON(uri)}})
	}()
	waitFreshness(t, rule.started)
	freshChange(s, uri, 1, "<?php ;;")
	close(rule.release)
	waitFreshness(t, done)
	if failure == nil || failure.Code != codeRequestFailed {
		t.Fatal(failure)
	}
}
