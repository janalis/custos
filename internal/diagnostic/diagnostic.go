// Package diagnostic defines findings, fixes and positioned results shared by adapters.
package diagnostic

import "custos/internal/php/syntax"

// Severity is the default reporting level of a rule.
type Severity string

const (
	SeverityError   Severity = "error"
	SeverityWarning Severity = "warning"
	SeverityInfo    Severity = "info"
)

// TextEdit replaces Span with NewText.
type TextEdit struct {
	Span    syntax.Span
	NewText string
}

// Fix is a lazily built quick-fix.
type Fix struct {
	Title string
	Edits func() []TextEdit
}

// Finding is one reported problem.
type Finding struct {
	Rule     string
	Severity Severity
	Span     syntax.Span
	Message  string
	Fixes    []Fix
}

// FileResult is the outcome for one file.
type FileResult struct {
	Path     string // as discovered (relative to the working directory when possible)
	Src      []byte
	Findings []Finding
	Errors   []syntax.Error // syntax errors
	Err      error          // I/O error
}

// Item is one finding with resolved positions (1-based line/column, rune columns).
type Item struct {
	Path      string `json:"path"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndLine   int    `json:"endLine"`
	EndColumn int    `json:"endColumn"`
	Rule      string `json:"rule"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Fixable   bool   `json:"fixable"`
}

// Items flattens results into positioned items (syntax errors use rule "syntax").
func Items(results []FileResult) []Item {
	var out []Item
	for _, r := range results {
		if r.Err != nil {
			out = append(out, Item{Path: r.Path, Line: 1, Column: 1, EndLine: 1, EndColumn: 1, Rule: "io", Severity: string(SeverityError), Message: r.Err.Error()})
			continue
		}
		if len(r.Findings) == 0 && len(r.Errors) == 0 {
			continue
		}
		li := syntax.NewLineIndex(r.Src)
		pos := func(s syntax.Span) (int, int, int, int) {
			l1, c1 := li.RuneColumn(s.Start)
			l2, c2 := li.RuneColumn(s.End)
			return l1 + 1, c1 + 1, l2 + 1, c2 + 1
		}
		for _, e := range r.Errors {
			l1, c1, l2, c2 := pos(e.Span)
			out = append(out, Item{Path: r.Path, Line: l1, Column: c1, EndLine: l2, EndColumn: c2, Rule: "syntax", Severity: string(SeverityError), Message: e.Msg})
		}
		for _, f := range r.Findings {
			l1, c1, l2, c2 := pos(f.Span)
			out = append(out, Item{Path: r.Path, Line: l1, Column: c1, EndLine: l2, EndColumn: c2, Rule: f.Rule, Severity: string(f.Severity), Message: f.Message, Fixable: len(f.Fixes) > 0})
		}
	}
	return out
}
