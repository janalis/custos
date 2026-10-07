package lsp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"custos/internal/analysis"
	"custos/internal/config"
	"custos/internal/fix"
	"custos/internal/index"
	"custos/internal/meta"
	"custos/internal/rules"
	"custos/internal/runner"
	"custos/internal/syntax"
)

// Debounce is the delay between the last edit and re-analysis.
var Debounce = 150 * time.Millisecond

const (
	cmdFixFile = "custos.fixFile"
	cmdFixRule = "custos.fixRule"
	kindFixAll = "source.fixAll.custos"
)

type document struct {
	uri      string
	path     string
	version  int
	text     []byte
	lines    *syntax.LineIndex
	findings []analysis.Finding // for analyzedVersion
	analyzed int                // version the findings belong to (-1 = none)
	timer    *time.Timer
}

// Server is the custos language server.
type Server struct {
	c        *conn
	mu       sync.Mutex
	docs     map[string]*document
	engine   *analysis.Engine
	cfg      *config.Config
	initOpts json.RawMessage
	root     string
	resolve  bool // client supports codeAction/resolve
	// watchDynamic: client supports dynamic registration of file watchers.
	watchDynamic bool
	// indexing is true while the initial project index is being built;
	// file changes arriving meanwhile are queued in pending and replayed.
	indexing bool
	pending  []fileChange
	index    *index.Index
	shutdown bool
	// sem bounds concurrent analyses.
	sem chan struct{}
}

// Serve runs the server on r/w until exit.
func Serve(ctx context.Context, r io.Reader, w io.Writer) error {
	s := &Server{c: newConn(r, w), docs: map[string]*document{}, sem: make(chan struct{}, runtime.GOMAXPROCS(0))}
	for {
		m, err := s.c.read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			var rerr *rpcError
			if errors.As(err, &rerr) {
				_ = s.c.reply(nil, nil, rerr)
				continue
			}
			return err
		}
		if m.Method == "" {
			continue // response to one of our requests
		}
		if m.Method == "exit" {
			return nil
		}
		result, rerr := s.handle(m)
		if m.ID != nil {
			if err := s.c.reply(m.ID, result, rerr); err != nil {
				return err
			}
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
	}
}

func (s *Server) handle(m *message) (result any, rerr *rpcError) {
	// A crashing rule must never take the server down; a request that
	// crashed gets an error, not a null success.
	defer s.guard("handling "+m.Method, func() {
		result, rerr = nil, &rpcError{Code: codeInternalError, Message: "custos crashed handling " + m.Method}
	})
	switch m.Method {
	case "initialize":
		return s.initialize(m.Params)
	case "initialized":
		if s.watchDynamic {
			// Ask the client to report PHP file changes so the project index
			// follows edits made outside open buffers (git checkout, codegen…).
			_ = s.c.request("client/registerCapability", map[string]any{"registrations": []any{map[string]any{
				"id": "custos-php-watch", "method": "workspace/didChangeWatchedFiles",
				"registerOptions": map[string]any{"watchers": []any{map[string]any{"globPattern": "**/*.php"}}},
			}}})
		}
		if s.beginIndexing() {
			go s.buildIndex()
		}
		return nil, nil
	case "workspace/didChangeWatchedFiles":
		var p struct {
			Changes []fileChange `json:"changes"`
		}
		_ = json.Unmarshal(m.Params, &p)
		go func() {
			defer s.guard("file changes", nil)
			s.watchedFilesChanged(p.Changes)
		}()
		return nil, nil
	case "textDocument/didSave":
		var p struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
		}
		_ = json.Unmarshal(m.Params, &p)
		s.reindexDoc(p.TextDocument.URI)
		return nil, nil
	case "$/cancelRequest", "$/setTrace":
		return nil, nil
	case "shutdown":
		s.shutdown = true
		return nil, nil
	case "textDocument/didOpen":
		var p didOpenParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		s.open(p)
		return nil, nil
	case "textDocument/didChange":
		var p didChangeParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		s.change(p)
		return nil, nil
	case "textDocument/didClose":
		var p struct {
			TextDocument struct {
				URI string `json:"uri"`
			} `json:"textDocument"`
		}
		_ = json.Unmarshal(m.Params, &p)
		s.close(p.TextDocument.URI)
		return nil, nil
	case "workspace/didChangeConfiguration":
		var p struct {
			Settings json.RawMessage `json:"settings"`
		}
		_ = json.Unmarshal(m.Params, &p)
		if len(p.Settings) > 0 && string(p.Settings) != "null" {
			s.mu.Lock()
			s.initOpts = p.Settings
			s.mu.Unlock()
		}
		if err := s.configure(); err != nil {
			s.logf("configuration: %v", err)
		}
		s.reanalyzeAll()
		return nil, nil
	case "textDocument/codeAction":
		var p codeActionParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.codeActions(p), nil
	case "codeAction/resolve":
		var a codeAction
		if err := json.Unmarshal(m.Params, &a); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.resolveAction(a)
	case "workspace/executeCommand":
		var p executeCommandParams
		if err := json.Unmarshal(m.Params, &p); err != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
		}
		return s.executeCommand(p)
	}
	if m.ID != nil {
		return nil, &rpcError{Code: codeMethodNotFound, Message: "method not found: " + m.Method}
	}
	return nil, nil
}

// ---- lifecycle ---------------------------------------------------------------------

func (s *Server) initialize(raw json.RawMessage) (any, *rpcError) {
	var p initializeParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
	}
	s.root = "."
	switch {
	case len(p.WorkspaceFolders) > 0:
		s.root = uriToPath(p.WorkspaceFolders[0].URI)
	case p.RootURI != "":
		s.root = uriToPath(p.RootURI)
	case p.RootPath != "":
		s.root = p.RootPath
	}
	s.initOpts = p.InitializationOptions
	s.resolve = p.Capabilities.TextDocument.CodeAction.ResolveSupport != nil
	s.watchDynamic = p.Capabilities.Workspace.DidChangeWatchedFiles.DynamicRegistration
	if err := s.configure(); err != nil {
		return nil, &rpcError{Code: codeRequestFailed, Message: err.Error()}
	}
	return map[string]any{
		"capabilities": map[string]any{
			"positionEncoding": "utf-16",
			"textDocumentSync": map[string]any{"openClose": true, "change": 2, "save": map[string]any{"includeText": false}},
			"codeActionProvider": map[string]any{
				"codeActionKinds": []string{"quickfix", kindFixAll},
				"resolveProvider": true,
			},
			"executeCommandProvider": map[string]any{"commands": []string{cmdFixFile, cmdFixRule}},
		},
		"serverInfo": map[string]any{"name": "custos", "version": Version},
	}, nil
}

// registry returns the rules served (overridable in tests).
var registry = rules.All

// Version is reported in serverInfo (set by main).
var Version = "dev"

// configure (re)builds the engine from custos.json + initializationOptions.
func (s *Server) configure() error {
	cfg, err := config.Load(s.root)
	if err != nil {
		return err
	}
	if len(s.initOpts) > 0 && string(s.initOpts) != "null" {
		var over config.File
		if err := json.Unmarshal(s.initOpts, &over); err != nil {
			return fmt.Errorf("initializationOptions: %w", err)
		}
		if cfg, err = mergeOverrides(cfg, over); err != nil {
			return err
		}
	}
	e, err := analysis.NewEngine(registry(), cfg.Analysis())
	if err != nil {
		return err
	}
	s.mu.Lock()
	if s.index != nil {
		e.SetIndex(s.index)
	}
	s.cfg, s.engine = cfg, e
	s.mu.Unlock()
	return nil
}

// mergeOverrides applies client-provided settings on top of the project config.
func mergeOverrides(base *config.Config, over config.File) (*config.Config, error) {
	f := config.File{PHP: over.PHP, ComparisonStyle: over.ComparisonStyle, ShortOpenTag: over.ShortOpenTag, Rules: over.Rules}
	if f.PHP == "" {
		f.PHP = base.PHP.String()
	}
	if f.ComparisonStyle == "" && base.ComparisonStyle == analysis.StyleYoda {
		f.ComparisonStyle = "yoda"
	}
	if f.ShortOpenTag == nil {
		f.ShortOpenTag = &base.ShortOpenTag
	}
	merged, err := config.Resolve(base.Root, f)
	if err != nil {
		return nil, err
	}
	// What to analyse stays the project's choice.
	merged.Paths, merged.Exclude, merged.Baseline = base.Paths, base.Exclude, base.Baseline
	for id, rc := range base.Rules {
		if _, ok := merged.Rules[id]; !ok {
			merged.Rules[id] = rc
		}
	}
	return merged, nil
}

// guard, deferred, recovers a panic: it is logged and onPanic (if any) runs.
func (s *Server) guard(what string, onPanic func()) {
	if r := recover(); r != nil {
		s.logf("%s panicked: %v", what, r)
		if onPanic != nil {
			onPanic()
		}
	}
}

func (s *Server) logf(format string, args ...any) {
	_ = s.c.notify("window/logMessage", map[string]any{"type": 3, "message": "custos: " + fmt.Sprintf(format, args...)})
}

// ---- project index -----------------------------------------------------------------

// beginIndexing reports whether an enabled rule needs cross-file symbols
// and, if so, marks the index as being built: file changes arriving from
// now on are queued for buildIndex instead of being dropped.
func (s *Server) beginIndexing() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.engine == nil || !s.engine.NeedsIndex() {
		return false
	}
	s.indexing = true
	return true
}

// buildIndex indexes the workspace (and vendor) after beginIndexing, then
// re-analyses open documents.
func (s *Server) buildIndex() {
	s.mu.Lock()
	e, cfg := s.engine, s.cfg
	s.mu.Unlock()
	start := time.Now()
	var paths []string
	for _, p := range cfg.Paths {
		paths = append(paths, filepath.Join(cfg.Root, p))
	}
	files, err := runner.Discover(paths, cfg.Exclude)
	if err != nil {
		s.logf("index: %v", err)
		s.mu.Lock()
		s.indexing, s.pending = false, nil
		s.mu.Unlock()
		return
	}
	ix := runner.BuildIndex(runner.IndexSources(cfg.Root, files), syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag})
	s.mu.Lock()
	if s.engine == e {
		e.SetIndex(ix)
		s.index = ix
	}
	s.indexing = false
	queued := s.pending
	s.pending = nil
	s.mu.Unlock()
	if len(queued) > 0 {
		s.watchedFilesChanged(queued) // replays and re-analyses
	}
	n, _, _, _ := ix.Stats()
	s.logf("indexed %d files in %v", n, time.Since(start).Round(time.Millisecond))
	s.reanalyzeAll()
}

// fileChange is one workspace/didChangeWatchedFiles event.
type fileChange struct {
	URI  string `json:"uri"`
	Type int    `json:"type"` // 1 created, 2 changed, 3 deleted
}

// watchedFilesChanged updates the project index for files changed on disk
// and re-analyses open documents (their semantic findings may depend on them).
// Changes arriving while the initial index is built are queued and replayed.
func (s *Server) watchedFilesChanged(changes []fileChange) {
	s.mu.Lock()
	if s.indexing {
		s.pending = append(s.pending, changes...)
		s.mu.Unlock()
		return
	}
	ix, cfg := s.index, s.cfg
	s.mu.Unlock()
	if ix == nil || len(changes) == 0 {
		return
	}
	touched := false
	for _, ch := range changes {
		path := uriToPath(ch.URI)
		if !strings.HasSuffix(strings.ToLower(path), ".php") {
			continue
		}
		touched = true
		if ch.Type == 3 {
			ix.Remove(path)
			continue
		}
		src, err := runner.ReadSource(path)
		if err != nil {
			ix.Remove(path)
			continue
		}
		fs := runner.ExtractSymbols(path, src, syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag})
		ix.DropStaleInferred(fs)
		ix.Add(fs)
	}
	if touched {
		s.reanalyzeAll()
	}
}

// reindexDoc refreshes the saved document's symbols in the project index.
func (s *Server) reindexDoc(uri string) {
	s.mu.Lock()
	d, ok := s.docs[uri]
	ix, cfg := s.index, s.cfg
	var text []byte
	var path string
	if ok {
		text, path = d.text, d.path
	}
	s.mu.Unlock()
	if !ok || ix == nil {
		return
	}
	fs := runner.ExtractSymbols(path, text, syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag})
	ix.DropStaleInferred(fs)
	ix.Add(fs)
}

// ---- documents -----------------------------------------------------------------------

func (s *Server) open(p didOpenParams) {
	if !isPHP(p.TextDocument) {
		return
	}
	d := &document{uri: p.TextDocument.URI, path: uriToPath(p.TextDocument.URI), version: p.TextDocument.Version, text: []byte(p.TextDocument.Text), analyzed: -1}
	d.lines = syntax.NewLineIndex(d.text)
	s.mu.Lock()
	s.docs[d.uri] = d
	s.mu.Unlock()
	s.analyzeNow(d.uri)
}

func isPHP(td textDocumentItem) bool {
	return td.LanguageID == "php" || strings.HasSuffix(strings.ToLower(td.URI), ".php") || td.LanguageID == ""
}

func (s *Server) change(p didChangeParams) {
	s.mu.Lock()
	d, ok := s.docs[p.TextDocument.URI]
	if !ok {
		s.mu.Unlock()
		return
	}
	for _, ch := range p.ContentChanges {
		if ch.Range == nil {
			d.text = []byte(ch.Text)
		} else {
			start := d.lines.OffsetUTF16(ch.Range.Start.Line, ch.Range.Start.Character)
			end := d.lines.OffsetUTF16(ch.Range.End.Line, ch.Range.End.Character)
			if end < start {
				start, end = end, start
			}
			buf := make([]byte, 0, len(d.text)-int(end-start)+len(ch.Text))
			buf = append(buf, d.text[:start]...)
			buf = append(buf, ch.Text...)
			buf = append(buf, d.text[end:]...)
			d.text = buf
		}
		d.lines = syntax.NewLineIndex(d.text)
	}
	d.version = p.TextDocument.Version
	if d.timer != nil {
		d.timer.Stop()
	}
	uri := d.uri
	d.timer = time.AfterFunc(Debounce, func() { s.analyzeNow(uri) })
	s.mu.Unlock()
}

func (s *Server) close(uri string) {
	s.mu.Lock()
	if d, ok := s.docs[uri]; ok && d.timer != nil {
		d.timer.Stop()
	}
	delete(s.docs, uri)
	s.mu.Unlock()
	_ = s.c.notify("textDocument/publishDiagnostics", publishDiagnosticsParams{URI: uri, Diagnostics: []diagnostic{}})
}

func (s *Server) reanalyzeAll() {
	s.mu.Lock()
	uris := make([]string, 0, len(s.docs))
	for u := range s.docs {
		uris = append(uris, u)
	}
	s.mu.Unlock()
	for _, u := range uris {
		s.analyzeNow(u)
	}
}

// analyzeNow analyses the current text of uri and publishes diagnostics if
// the document did not change meanwhile.
func (s *Server) analyzeNow(uri string) {
	s.mu.Lock()
	d, ok := s.docs[uri]
	if !ok || s.engine == nil {
		s.mu.Unlock()
		return
	}
	text, version, path, e, cfg := d.text, d.version, d.path, s.engine, s.cfg
	s.mu.Unlock()

	s.sem <- struct{}{}
	findings := func() []analysis.Finding {
		defer s.guard("analysis of "+path, nil)
		f := syntax.ParseBest(path, text, syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag})
		return e.Analyze(f)
	}()
	<-s.sem

	s.mu.Lock()
	d, ok = s.docs[uri]
	if !ok || d.version != version {
		s.mu.Unlock()
		return // stale
	}
	d.findings, d.analyzed = findings, version
	lines := d.lines
	s.mu.Unlock()

	diags := make([]diagnostic, 0, len(findings))
	for _, f := range findings {
		diags = append(diags, toDiagnostic(lines, f))
	}
	v := version
	_ = s.c.notify("textDocument/publishDiagnostics", publishDiagnosticsParams{URI: uri, Version: &v, Diagnostics: diags})
}

func toRange(lines *syntax.LineIndex, sp syntax.Span) lspRange {
	l1, c1 := lines.UTF16Column(sp.Start)
	l2, c2 := lines.UTF16Column(sp.End)
	return lspRange{Start: position{l1, c1}, End: position{l2, c2}}
}

func toDiagnostic(lines *syntax.LineIndex, f analysis.Finding) diagnostic {
	sev := 3
	switch f.Severity {
	case meta.SeverityError:
		sev = 1
	case meta.SeverityWarning:
		sev = 2
	}
	return diagnostic{Range: toRange(lines, f.Span), Severity: sev, Code: f.Rule, Source: "custos", Message: f.Message,
		Data: &diagData{Rule: f.Rule, Start: f.Span.Start, End: f.Span.End}}
}

// ---- code actions ---------------------------------------------------------------------

// actionData identifies a fix for lazy resolution.
type actionData struct {
	URI     string `json:"uri"`
	Version int    `json:"version"`
	Rule    string `json:"rule,omitempty"`
	Start   uint32 `json:"start,omitempty"`
	End     uint32 `json:"end,omitempty"`
	Fix     int    `json:"fix,omitempty"`
	All     bool   `json:"all,omitempty"`
}

func (s *Server) codeActions(p codeActionParams) []codeAction {
	s.mu.Lock()
	d, ok := s.docs[p.TextDocument.URI]
	if !ok || d.analyzed != d.version {
		s.mu.Unlock()
		return []codeAction{}
	}
	lines, findings, version := d.lines, d.findings, d.version
	s.mu.Unlock()

	want := func(kind string) bool {
		if len(p.Context.Only) == 0 {
			return true
		}
		for _, k := range p.Context.Only {
			if kind == k || strings.HasPrefix(kind, k+".") {
				return true
			}
		}
		return false
	}
	start := lines.OffsetUTF16(p.Range.Start.Line, p.Range.Start.Character)
	end := lines.OffsetUTF16(p.Range.End.Line, p.Range.End.Character)
	actions := []codeAction{}
	fixable := 0
	// Diagnostics the client sends back identify findings exactly via data.
	named := map[diagData]bool{}
	for _, d := range p.Context.Diagnostics {
		if d.Data != nil {
			named[*d.Data] = true
		}
	}
	for _, f := range findings {
		if len(f.Fixes) == 0 {
			continue
		}
		fixable++
		inRange := f.Span.End >= start && f.Span.Start <= end
		if !want("quickfix") || (!inRange && !named[diagData{Rule: f.Rule, Start: f.Span.Start, End: f.Span.End}]) {
			continue
		}
		diag := toDiagnostic(lines, f)
		for i, fx := range f.Fixes {
			data := actionData{URI: p.TextDocument.URI, Version: version, Rule: f.Rule, Start: f.Span.Start, End: f.Span.End, Fix: i}
			a := codeAction{Title: fx.Title, Kind: "quickfix", Diagnostics: []diagnostic{diag}, IsPreferred: i == 0, Data: mustJSON(data)}
			if !s.resolve {
				a.Edit = s.buildEdit(data)
			}
			actions = append(actions, a)
		}
	}
	if fixable > 0 && want(kindFixAll) {
		data := actionData{URI: p.TextDocument.URI, Version: version, All: true}
		a := codeAction{Title: "custos: fix all problems in file", Kind: kindFixAll, Data: mustJSON(data)}
		if !s.resolve {
			a.Edit = s.buildEdit(data)
		}
		actions = append(actions, a)
	}
	return actions
}

func (s *Server) resolveAction(a codeAction) (any, *rpcError) {
	var data actionData
	if err := json.Unmarshal(a.Data, &data); err != nil {
		return nil, &rpcError{Code: codeInvalidParams, Message: "bad code action data"}
	}
	a.Edit = s.buildEdit(data)
	if a.Edit == nil {
		return nil, &rpcError{Code: codeRequestFailed, Message: "document changed; fix no longer applies"}
	}
	return a, nil
}

// buildEdit computes the workspace edit for a fix (or fix-all).
func (s *Server) buildEdit(data actionData) *workspaceEdit {
	s.mu.Lock()
	d, ok := s.docs[data.URI]
	if !ok || d.version != data.Version || d.analyzed != d.version {
		s.mu.Unlock()
		return nil
	}
	text, lines, findings, path, e, cfg := d.text, d.lines, d.findings, d.path, s.engine, s.cfg
	s.mu.Unlock()

	var edits []analysis.TextEdit
	if data.All {
		res := fix.FixSource(e, path, text, fix.Options{Parse: syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}})
		if res.Applied == 0 {
			return nil
		}
		edits = []analysis.TextEdit{{Span: syntax.Span{Start: 0, End: uint32(len(text))}, NewText: string(res.Source)}}
	} else {
		for _, f := range findings {
			if f.Rule == data.Rule && f.Span.Start == data.Start && f.Span.End == data.End && data.Fix >= 0 && data.Fix < len(f.Fixes) {
				edits = f.Fixes[data.Fix].Edits()
				break
			}
		}
		if edits == nil {
			return nil
		}
	}
	return docEdit(data.URI, data.Version, lines, edits)
}

func docEdit(uri string, version int, lines *syntax.LineIndex, edits []analysis.TextEdit) *workspaceEdit {
	tes := make([]textEdit, 0, len(edits))
	for _, e := range edits {
		tes = append(tes, textEdit{Range: toRange(lines, e.Span), NewText: e.NewText})
	}
	v := version
	return &workspaceEdit{DocumentChanges: []textDocumentEdit{{TextDocument: versionedTextDocumentIdentifier{URI: uri, Version: &v}, Edits: tes}}}
}

func (s *Server) executeCommand(p executeCommandParams) (any, *rpcError) {
	if len(p.Arguments) == 0 {
		return nil, &rpcError{Code: codeInvalidParams, Message: "missing document URI argument"}
	}
	var uri string
	if err := json.Unmarshal(p.Arguments[0], &uri); err != nil {
		return nil, &rpcError{Code: codeInvalidParams, Message: "first argument must be a document URI"}
	}
	var only string
	if p.Command == cmdFixRule {
		if len(p.Arguments) < 2 || json.Unmarshal(p.Arguments[1], &only) != nil {
			return nil, &rpcError{Code: codeInvalidParams, Message: "custos.fixRule needs (uri, ruleId)"}
		}
	} else if p.Command != cmdFixFile {
		return nil, &rpcError{Code: codeMethodNotFound, Message: "unknown command " + p.Command}
	}
	s.mu.Lock()
	d, ok := s.docs[uri]
	if !ok {
		s.mu.Unlock()
		return nil, &rpcError{Code: codeInvalidParams, Message: "document not open"}
	}
	text, lines, version, path, e, cfg := d.text, d.lines, d.version, d.path, s.engine, s.cfg
	s.mu.Unlock()
	opt := fix.Options{Parse: syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}}
	if only != "" {
		m, ok := meta.Lookup(only)
		if !ok {
			return nil, &rpcError{Code: codeInvalidParams, Message: "unknown rule " + only}
		}
		opt.Filter = func(f analysis.Finding) bool { return f.Rule == m.ID }
	}
	res := fix.FixSource(e, path, text, opt)
	if res.Applied == 0 {
		return nil, nil
	}
	edit := docEdit(uri, version, lines, []analysis.TextEdit{{Span: syntax.Span{Start: 0, End: uint32(len(text))}, NewText: string(res.Source)}})
	if err := s.c.request("workspace/applyEdit", map[string]any{"label": "custos fix", "edit": edit}); err != nil {
		return nil, &rpcError{Code: codeInternalError, Message: err.Error()}
	}
	return nil, nil
}

// ---- helpers ------------------------------------------------------------------------

func mustJSON(v any) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}

func uriToPath(uri string) string {
	u, err := url.Parse(uri)
	if err != nil || u.Scheme != "file" {
		return uri
	}
	p := u.Path
	// file:///C:/dir → C:/dir (Windows drive letters).
	if len(p) >= 3 && p[0] == '/' && p[2] == ':' && ((p[1] >= 'a' && p[1] <= 'z') || (p[1] >= 'A' && p[1] <= 'Z')) {
		p = p[1:]
	}
	return filepath.FromSlash(p)
}
