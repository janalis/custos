//go:build unix

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Features is bench/features.json: what each compared tool offers, checked
// against its own documentation.
type Features struct {
	// Checked is the date the cells were last checked (YYYY-MM-DD).
	Checked  string    `json:"checked"`
	Tools    []FTool   `json:"tools"`
	Features []Feature `json:"features"`
}

// FTool is a column of the comparison.
type FTool struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	URL   string `json:"url"`
}

// Feature is a row: one cell per tool, by tool ID.
type Feature struct {
	Name  string          `json:"name"`
	Cells map[string]Cell `json:"cells"`
}

// Cell is what one tool offers for a feature: Value is "yes", "no",
// "partial" or "" (Note alone, such as a language name); Note qualifies it
// and Source links the documentation that says so.
type Cell struct {
	Value  string `json:"value,omitempty"`
	Note   string `json:"note,omitempty"`
	Source string `json:"source,omitempty"`
}

var (
	cellMarks    = map[string]string{"yes": "✓", "no": "—", "partial": "partial", "": ""}
	compactMarks = map[string]string{"yes": "✓", "no": "—", "partial": "◐"}
)

func loadFeatures(path string) (*Features, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var f Features
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&f); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if err := f.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return &f, nil
}

// validate requires a cell for every tool in every row, and nothing else.
func (f *Features) validate() error {
	if f.Checked == "" || len(f.Tools) == 0 {
		return fmt.Errorf("checked and tools are required")
	}
	ids := map[string]bool{}
	for _, t := range f.Tools {
		ids[t.ID] = true
	}
	for _, feat := range f.Features {
		for id, c := range feat.Cells {
			if !ids[id] {
				return fmt.Errorf("feature %q: unknown tool %q", feat.Name, id)
			}
			if _, ok := cellMarks[c.Value]; !ok || c.Value == "" && c.Note == "" {
				return fmt.Errorf("feature %q, tool %s: value must be yes, no or partial, or a note alone", feat.Name, id)
			}
		}
		for _, t := range f.Tools {
			if _, ok := feat.Cells[t.ID]; !ok {
				return fmt.Errorf("feature %q: no cell for %s", feat.Name, t.ID)
			}
		}
	}
	return nil
}

// table renders the comparison as a Markdown table followed by a line
// saying when it was checked. Text is escaped for both GitHub and the
// documentation site.
func (f *Features) table() string { return f.render(false) }

// compact renders the comparison with marks only, for the README: notes and
// sources stay on the benchmark page.
func (f *Features) compact() string { return f.render(true) }

func (f *Features) render(compact bool) string {
	var b strings.Builder
	b.WriteString("| Feature |")
	for _, t := range f.Tools {
		fmt.Fprintf(&b, " [%s](%s) |", escape(t.Label), t.URL)
	}
	b.WriteString("\n| :--- |")
	for range f.Tools {
		b.WriteString(" :---: |")
	}
	b.WriteString("\n")
	for _, feat := range f.Features {
		fmt.Fprintf(&b, "| %s |", escape(feat.Name))
		for _, t := range f.Tools {
			c := feat.Cells[t.ID]
			if compact && c.Value != "" {
				fmt.Fprintf(&b, " %s |", compactMarks[c.Value])
				continue
			}
			sep := " "
			if c.Value == "partial" && c.Note != "" {
				sep = ": "
			}
			text := strings.TrimSpace(cellMarks[c.Value] + sep + escape(c.Note))
			if c.Source != "" {
				text = fmt.Sprintf("[%s](%s)", text, c.Source)
			}
			fmt.Fprintf(&b, " %s |", text)
		}
		b.WriteString("\n")
	}
	if compact {
		return b.String()
	}
	fmt.Fprintf(&b, "\nChecked against each tool's documentation on %s; linked cells cite it.\n", f.Checked)
	return b.String()
}
