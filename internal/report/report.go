// Package report renders analysis results (text, json, checkstyle, github).
package report

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"custos/internal/meta"
	"custos/internal/runner"
	"custos/internal/syntax"
)

// Formats lists the supported output formats.
var Formats = []string{"text", "json", "checkstyle", "github", "sarif"}

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
func Items(results []runner.FileResult) []Item {
	var out []Item
	for _, r := range results {
		if r.Err != nil {
			out = append(out, Item{Path: r.Path, Line: 1, Column: 1, EndLine: 1, EndColumn: 1, Rule: "io", Severity: string(meta.SeverityError), Message: r.Err.Error()})
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
			out = append(out, Item{Path: r.Path, Line: l1, Column: c1, EndLine: l2, EndColumn: c2, Rule: "syntax", Severity: string(meta.SeverityError), Message: e.Msg})
		}
		for _, f := range r.Findings {
			l1, c1, l2, c2 := pos(f.Span)
			out = append(out, Item{Path: r.Path, Line: l1, Column: c1, EndLine: l2, EndColumn: c2, Rule: f.Rule, Severity: string(f.Severity), Message: f.Message, Fixable: len(f.Fixes) > 0})
		}
	}
	return out
}

// Write renders items in the given format.
func Write(w io.Writer, format string, items []Item, files int) error {
	switch format {
	case "text":
		return writeText(w, items, files)
	case "json":
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		if items == nil {
			items = []Item{}
		}
		return enc.Encode(map[string]any{"files": files, "findings": items})
	case "checkstyle":
		return writeCheckstyle(w, items)
	case "sarif":
		return writeSARIF(w, items)
	case "github":
		for _, it := range items {
			level := map[string]string{"error": "error", "warning": "warning"}[it.Severity]
			if level == "" {
				level = "notice"
			}
			fmt.Fprintf(w, "::%s file=%s,line=%d,col=%d,endLine=%d,endColumn=%d,title=%s::%s\n",
				level, it.Path, it.Line, it.Column, it.EndLine, it.EndColumn, it.Rule, escapeGitHub(it.Message))
		}
		return nil
	}
	return fmt.Errorf("unknown format %q (want one of %s)", format, strings.Join(Formats, ", "))
}

func writeText(w io.Writer, items []Item, files int) error {
	counts := map[string]int{}
	for _, it := range items {
		fix := ""
		if it.Fixable {
			fix = " (fixable)"
		}
		fmt.Fprintf(w, "%s:%d:%d: %s: %s [%s]%s\n", it.Path, it.Line, it.Column, it.Severity, it.Message, it.Rule, fix)
		counts[it.Severity]++
	}
	_, err := fmt.Fprintf(w, "\n%d file(s) analysed: %d error(s), %d warning(s), %d info\n", files, counts["error"], counts["warning"], counts["info"])
	return err
}

func escapeGitHub(s string) string {
	return strings.NewReplacer("%", "%25", "\r", "%0D", "\n", "%0A").Replace(s)
}

type csError struct {
	Line     int    `xml:"line,attr"`
	Column   int    `xml:"column,attr"`
	Severity string `xml:"severity,attr"`
	Message  string `xml:"message,attr"`
	Source   string `xml:"source,attr"`
}

type csFile struct {
	Name   string    `xml:"name,attr"`
	Errors []csError `xml:"error"`
}

func writeCheckstyle(w io.Writer, items []Item) error {
	type root struct {
		XMLName xml.Name `xml:"checkstyle"`
		Version string   `xml:"version,attr"`
		Files   []csFile `xml:"file"`
	}
	r := root{Version: "4.3"}
	idx := map[string]int{}
	for _, it := range items {
		i, ok := idx[it.Path]
		if !ok {
			i = len(r.Files)
			idx[it.Path] = i
			r.Files = append(r.Files, csFile{Name: it.Path})
		}
		sev := it.Severity
		if sev == "warning" || sev == "error" {
		} else {
			sev = "info"
		}
		r.Files[i].Errors = append(r.Files[i].Errors, csError{Line: it.Line, Column: it.Column, Severity: sev, Message: it.Message, Source: "custos." + it.Rule})
	}
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	enc := xml.NewEncoder(w)
	enc.Indent("", "  ")
	if err := enc.Encode(r); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\n")
	return err
}

// writeSARIF emits SARIF 2.1.0 (GitHub code scanning and most CI viewers).
func writeSARIF(w io.Writer, items []Item) error {
	level := map[string]string{"error": "error", "warning": "warning", "info": "note"}
	ruleIdx := map[string]int{}
	var rules []map[string]any
	results := []map[string]any{}
	for _, it := range items {
		if _, ok := ruleIdx[it.Rule]; !ok {
			ruleIdx[it.Rule] = len(rules)
			r := map[string]any{"id": it.Rule}
			if m, ok := meta.Lookup(it.Rule); ok {
				r["properties"] = map[string]any{"tags": []string{m.Group}}
				if d, ok := meta.Describe(m.ID); ok && d.Summary != "" {
					r["shortDescription"] = map[string]any{"text": strings.SplitN(d.Summary, "\n\n", 2)[0]}
				}
			}
			rules = append(rules, r)
		}
		lv := level[it.Severity]
		if lv == "" {
			lv = "warning"
		}
		results = append(results, map[string]any{
			"ruleId": it.Rule, "ruleIndex": ruleIdx[it.Rule], "level": lv,
			"message": map[string]any{"text": it.Message},
			"locations": []any{map[string]any{"physicalLocation": map[string]any{
				"artifactLocation": map[string]any{"uri": filepathToURI(it.Path)},
				"region":           map[string]any{"startLine": it.Line, "startColumn": it.Column, "endLine": it.EndLine, "endColumn": it.EndColumn},
			}}},
		})
	}
	doc := map[string]any{
		"$schema": "https://json.schemastore.org/sarif-2.1.0.json",
		"version": "2.1.0",
		"runs": []any{map[string]any{
			"tool":    map[string]any{"driver": map[string]any{"name": "custos", "rules": rules}},
			"results": results,
		}},
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}

func filepathToURI(p string) string {
	return strings.ReplaceAll(strings.TrimPrefix(p, "./"), "\\", "/")
}
