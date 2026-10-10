package lsp

import (
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
	d.version = p.TextDocument.Version
	if d.timer != nil {
		d.timer.Stop()
	}
	uri := d.uri
	d.timer = time.AfterFunc(Debounce, func() { s.analyzeChanged(uri) })
	s.mu.Unlock()
}

func (s *Server) close(uri string) {
	s.mu.Lock()
	if d, ok := s.docs[uri]; ok && d.timer != nil {
		d.timer.Stop()
	}
	delete(s.docs, uri)
	s.mu.Unlock()
	_ = s.c.notify("textDocument/publishDiagnostics", publishDiagnosticsParams{URI: uri, Diagnostics: []protocolDiagnostic{}})
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

// analyzeChanged republishes buffer summaries before checking dependent documents.
func (s *Server) analyzeChanged(uri string) {
	if s.flowEnabled() && s.reindexDoc(uri) {
		s.reanalyzeAll()
		return
	}
	s.analyzeNow(uri)
}
