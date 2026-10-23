package lsp

import (
	"maps"
	"slices"
	"strings"
	"time"

	"custos/internal/php/syntax"
)

func (s *Server) open(p didOpenParams) {
	if !isPHP(p.TextDocument) {
		return
	}
	d := &document{uri: p.TextDocument.URI, path: uriToPath(p.TextDocument.URI), version: p.TextDocument.Version, text: []byte(p.TextDocument.Text), analyzed: -1}
	d.lines = syntax.NewLineIndex(d.text)
	s.mu.Lock()
	s.docs[d.uri] = d
	s.mu.Unlock()
	s.analyzeChanged(d.uri)
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
	d.revision++
	d.analyzed = -1
	d.version = p.TextDocument.Version
	if d.timer != nil {
		d.timer.Stop()
	}
	uri := d.uri
	d.timer = time.AfterFunc(Debounce, func() { s.analyzeChanged(uri) })
	s.mu.Unlock()
}

func (s *Server) close(uri string) {
	func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if d, ok := s.docs[uri]; ok && d.timer != nil {
			d.timer.Stop()
		}
		delete(s.docs, uri)
		_ = s.c.notify("textDocument/publishDiagnostics", publishDiagnosticsParams{URI: uri, Diagnostics: []protocolDiagnostic{}})
	}()
	if s.flowEnabled() {
		s.watchedFilesChanged([]fileChange{{URI: uri, Type: 2}})
	}
}

// flowEnabled reads the enabled rule contract under the server lock.
func (s *Server) flowEnabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.engine != nil && s.engine.NeedsFlow()
}

// analyzeChanged publishes the diagnostics of a changed document against the
// summaries of the other files, then has its own updated (see reflowLater):
// updating them may take seconds when the document lies on a cycle of
// calls, during which no declaration another file sees has changed.
func (s *Server) analyzeChanged(uri string) {
	if !s.flowEnabled() {
		s.analyzeNow(uri)
		return
	}
	s.indexDoc(uri)
	s.analyzeNow(uri)
	s.reflowLater(uri)
}

// reflowLater updates the flow summaries of uri's buffer once editing
// pauses for FlowDelay, with those of every document edited meanwhile: an
// update within a recursion keeps cores busy for seconds, which a pause
// after each keystroke would repeat.
func (s *Server) reflowLater(uri string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.flowPending == nil {
		s.flowPending = map[string]bool{}
	}
	s.flowPending[uri] = true
	if s.flowTimer != nil {
		s.flowTimer.Stop()
	}
	s.flowTimer = time.AfterFunc(FlowDelay, s.reflowPending)
}

// reflowPending updates the flow summaries of the documents reflowLater
// queued.
func (s *Server) reflowPending() {
	s.mu.Lock()
	uris := slices.Sorted(maps.Keys(s.flowPending))
	clear(s.flowPending)
	s.mu.Unlock()
	s.reflow(uris...)
}

// reflow updates the flow summaries of the buffers of uris (see
// reindexDoc), one update at a time: a document changed while one runs is
// queued, and its latest text summarized once that update ends. Open
// documents are analysed again when a summary changed.
func (s *Server) reflow(uris ...string) {
	s.mu.Lock()
	if s.reflowQueue == nil {
		s.reflowQueue = map[string]bool{}
	}
	for _, uri := range uris {
		s.reflowQueue[uri] = true
	}
	if s.reflowing {
		s.mu.Unlock()
		return
	}
	s.reflowing = true
	s.mu.Unlock()
	defer s.guard("flow update", nil)
	defer func() {
		s.mu.Lock()
		s.reflowing = false
		s.mu.Unlock()
	}()
	changed := false
	for {
		s.mu.Lock()
		uris := slices.Sorted(maps.Keys(s.reflowQueue))
		clear(s.reflowQueue)
		s.mu.Unlock()
		if len(uris) == 0 {
			break
		}
		for _, u := range uris {
			changed = s.reindexDoc(u) || changed
		}
	}
	if changed {
		s.reanalyzeAll()
	}
}
