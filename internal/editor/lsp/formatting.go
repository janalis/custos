package lsp

import (
	"bytes"
	"path/filepath"
	"strings"

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
		if c.Start < lo {
			// The gap before the first line starts on the line above.
			var ok bool
			if c, ok = firstIndent(text, c, lo); !ok {
				continue
			}
		}
		if c.Report || c.End > hi || outside[c.Group] {
			continue
		}
		edits = append(edits, textEdit{Range: toRange(lines, syntax.Span{Start: c.Start, End: c.End}), NewText: c.New})
	}
	return edits
}

// firstIndent narrows a change of the gap that ends the line before lo,
// the start of the range's first line, to that line's indentation, so that
// formatting a range re-indents its first line. It reports false when the
// change is grouped or report-only, when the part from lo is more than
// indentation, when its new text does not end with a line's indentation,
// or when the indentation is already right.
func firstIndent(text []byte, c format.Change, lo uint32) (format.Change, bool) {
	nl := strings.LastIndexByte(c.New, '\n')
	if c.Group != 0 || c.Report || c.End < lo || nl < 0 {
		return c, false
	}
	// A range starting on a blank line re-indents the line after it.
	start := lo
	if i := bytes.LastIndexByte(text[lo:c.End], '\n'); i >= 0 {
		start += uint32(i) + 1
	}
	if !indentation(string(text[start:c.End])) || !indentation(c.New[nl+1:]) {
		return c, false
	}
	c.Start, c.New = start, c.New[nl+1:]
	return c, c.New != string(text[start:c.End])
}

// indentation reports whether s holds only spaces and tabs.
func indentation(s string) bool { return strings.Trim(s, " \t") == "" }

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
