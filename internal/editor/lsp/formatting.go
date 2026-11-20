package lsp

import (
	"path/filepath"

	"github.com/janalis/custos/internal/format"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/project/config"
)

// formattingParams are those of textDocument/formatting and, with Range,
// textDocument/rangeFormatting. The request's FormattingOptions (tab size,
// spaces) are ignored: the project's format configuration (custos.json,
// .editorconfig, other tools' files) decides the layout, so that the editor
// writes what custos format does.
type formattingParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	Range *lspRange `json:"range"`
}

// formatting returns the edits that lay out the open document as custos
// format would: all of them, or those touching the lines of p.Range. A
// document the server does not hold, a file format.exclude leaves out or
// one the formatter skips gets no edit.
func (s *Server) formatting(p formattingParams) []textEdit {
	edits := []textEdit{}
	s.mu.Lock()
	d, cfg := s.docs[p.TextDocument.URI], s.cfg
	var text []byte
	var lines *syntax.LineIndex
	var path string
	if d != nil {
		// Edits replace the text and its index, never modify them.
		text, lines, path = d.text, d.lines, d.path
	}
	s.mu.Unlock()
	if d == nil || cfg == nil || cfg.FormatExcluded(path) {
		return edits
	}
	settings, err := cfg.FormatSettings(path, config.FormatFlags{})
	if err != nil {
		s.logf("format: %v", err)
		return edits
	}
	res := format.Source(path, text, format.Options{Settings: &settings.Settings, PHP: cfg.PHP})
	if res.Skipped != "" {
		s.logf("%s: %s, not formatted", filepath.Base(path), res.Skipped)
		return edits
	}
	lo, hi := uint32(0), uint32(len(text))
	if r := p.Range; r != nil {
		lo, hi = lineBounds(lines, r, hi)
	}
	// Only changes inside the lines, and a group of changes that only work
	// together only when all of it is inside.
	outside := map[int]bool{}
	for _, c := range res.Changes {
		if c.Group != 0 && (c.Start < lo || c.End > hi) {
			outside[c.Group] = true
		}
	}
	for _, c := range res.Changes {
		if c.Report || c.Start < lo || c.End > hi || outside[c.Group] {
			continue
		}
		edits = append(edits, textEdit{Range: toRange(lines, syntax.Span{Start: c.Start, End: c.End}), NewText: c.New})
	}
	return edits
}

// lineBounds returns the byte bounds of the whole lines r covers in a text
// of size bytes; a range ending at the start of a line leaves that line out.
func lineBounds(lines *syntax.LineIndex, r *lspRange, size uint32) (lo, hi uint32) {
	first, last := r.Start.Line, r.End.Line
	if last < first {
		first, last = last, first
	}
	if r.End.Character == 0 && last > first {
		last--
	}
	lo = lines.LineSpan(first).Start
	if first >= lines.Lines() {
		lo = size
	}
	hi = size
	if last < lines.Lines() {
		hi = lines.LineSpan(last).End
	}
	return lo, hi
}
