// Package baseline records known findings so that only new ones are
// reported. Entries are keyed by file path (relative to the baseline's
// directory), rule, message and the trimmed text of the flagged line, so
// they survive unrelated edits that move code around; a count per key
// tolerates duplicates.
package baseline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"custos/internal/report"
	"custos/internal/safeio"
	"custos/internal/syntax"
)

// Entry is one baseline key with an occurrence count.
type Entry struct {
	Path    string `json:"path"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
	Line    string `json:"line"` // trimmed source line of the finding start
	Count   int    `json:"count"`
}

// File is the on-disk format.
type File struct {
	Version int     `json:"version"`
	Entries []Entry `json:"entries"`
}

type key struct{ path, rule, msg, line string }

// Baseline is a loaded baseline.
type Baseline struct {
	dir    string
	counts map[key]int
}

// MaxFileSize bounds a baseline file (it may be named by the analysed
// repository's custos.json).
const MaxFileSize = 256 << 20

// Load reads a baseline file.
func Load(path string) (*Baseline, error) {
	b, err := safeio.ReadFile(path, MaxFileSize)
	if err != nil {
		return nil, err
	}
	var f File
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	abs, _ := filepath.Abs(path)
	bl := &Baseline{dir: filepath.Dir(abs), counts: map[key]int{}}
	for _, e := range f.Entries {
		bl.counts[key{e.Path, e.Rule, e.Message, e.Line}] += max(e.Count, 1)
	}
	return bl, nil
}

// Empty returns a baseline that suppresses nothing.
func Empty() *Baseline { return &Baseline{counts: map[key]int{}} }

// lineText returns the trimmed text of a 1-based line of the file at path.
func lineText(cache map[string][]string, path string, line int) string {
	lines, ok := cache[path]
	if !ok {
		// bounded like the analysis read (the path may name a device)
		b, _ := safeio.ReadPrefix(path, syntax.MaxFileSize+1)
		lines = strings.Split(string(bytes.ReplaceAll(b, []byte("\r\n"), []byte("\n"))), "\n")
		cache[path] = lines
	}
	if line >= 1 && line <= len(lines) {
		return strings.TrimSpace(lines[line-1])
	}
	return ""
}

func rel(dir, path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	if r, err := filepath.Rel(dir, abs); err == nil {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(abs)
}

// Filter removes items covered by the baseline and returns the rest plus
// the number of suppressed items.
func (b *Baseline) Filter(items []report.Item) ([]report.Item, int) {
	remaining := make(map[key]int, len(b.counts))
	for k, v := range b.counts {
		remaining[k] = v
	}
	cache := map[string][]string{}
	var out []report.Item
	suppressed := 0
	for _, it := range items {
		k := key{rel(b.dir, it.Path), it.Rule, it.Message, lineText(cache, it.Path, it.Line)}
		if remaining[k] > 0 {
			remaining[k]--
			suppressed++
			continue
		}
		out = append(out, it)
	}
	return out, suppressed
}

// Write creates a baseline at path covering items.
func Write(path string, items []report.Item) (int, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return 0, err
	}
	dir := filepath.Dir(abs)
	cache := map[string][]string{}
	counts := map[key]int{}
	for _, it := range items {
		if it.Rule == "syntax" || it.Rule == "internal" || it.Rule == "io" {
			continue
		}
		counts[key{rel(dir, it.Path), it.Rule, it.Message, lineText(cache, it.Path, it.Line)}]++
	}
	f := File{Version: 1}
	for k, n := range counts {
		f.Entries = append(f.Entries, Entry{Path: k.path, Rule: k.rule, Message: k.msg, Line: k.line, Count: n})
	}
	sort.Slice(f.Entries, func(i, j int) bool {
		a, b := f.Entries[i], f.Entries[j]
		if a.Path != b.Path {
			return a.Path < b.Path
		}
		if a.Rule != b.Rule {
			return a.Rule < b.Rule
		}
		return a.Line < b.Line
	})
	// strings and ints only: marshalling cannot fail
	data, _ := json.MarshalIndent(f, "", "  ")
	return len(f.Entries), os.WriteFile(path, append(data, '\n'), 0o644)
}
