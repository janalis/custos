package lsp

import (
	"custos/internal/diagnostic"
	"custos/internal/php/syntax"
	"custos/internal/project"
)

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
	findings := func() []diagnostic.Finding {
		defer s.guard("analysis of "+path, nil)
		return project.AnalyzeBuffer(e, path, text, syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}).Findings
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

	diags := make([]protocolDiagnostic, 0, len(findings))
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

func toDiagnostic(lines *syntax.LineIndex, f diagnostic.Finding) protocolDiagnostic {
	sev := 3
	switch f.Severity {
	case diagnostic.SeverityError:
		sev = 1
	case diagnostic.SeverityWarning:
		sev = 2
	}
	return protocolDiagnostic{
		Range: toRange(lines, f.Span), Severity: sev, Code: f.Rule, Source: "custos", Message: f.Message,
		Data: &diagData{Rule: f.Rule, Start: f.Span.Start, End: f.Span.End},
	}
}
