// Package baseline records known findings so that only new ones are
// reported. Entries are keyed by file path (relative to the baseline's
// directory), rule, message and the trimmed text of the flagged line, so
// they survive unrelated edits that move code around; a count per key
// tolerates duplicates. A finding that moved to a related rule (see
// related) stays covered by the entry of the rule it moved from.
package baseline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"custos/internal/diagnostic"
	"custos/internal/inspection/meta"
	"custos/internal/php/syntax"
	"custos/internal/platform/safeio"
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

// loc is a key without its message: where a finding is and which rule.
type loc struct{ path, rule, line string }

// Baseline is a loaded baseline.
type Baseline struct {
	dir    string
	counts map[key]int
	byLoc  map[loc][]key   // entries by location, for moved findings
	paths  map[string]bool // paths with at least one entry
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
	abs, _ := filepath.Abs(path)
	bl, err := Parse(b, filepath.Dir(abs))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return bl, nil
}

// Parse decodes a baseline whose entry paths are relative to dir. A
// relative dir matches the relative paths of findings against it as they
// are (the root-relative names of a captured snapshot, say).
func Parse(data []byte, dir string) (*Baseline, error) {
	var f File
	if err := json.Unmarshal(data, &f); err != nil {
		return nil, err
	}
	bl := &Baseline{dir: dir, counts: map[key]int{}, byLoc: map[loc][]key{}, paths: map[string]bool{}}
	for _, e := range f.Entries {
		k := key{e.Path, e.Rule, e.Message, e.Line}
		if bl.counts[k] == 0 {
			l := loc{e.Path, e.Rule, e.Line}
			bl.byLoc[l] = append(bl.byLoc[l], k)
		}
		bl.counts[k] += max(e.Count, 1)
		bl.paths[e.Path] = true
	}
	return bl, nil
}

// Empty returns a baseline that suppresses nothing.
func Empty() *Baseline { return &Baseline{counts: map[key]int{}} }

// related maps a rule to the rules its findings may have moved from or
// to: those it names in suppressedBy and those naming it there. A rule
// that leaves a finding to another (NullPointerException to
// PossiblyNullAccess, say) reported it before that rule existed or ran, so
// an entry recorded for one covers the other at the same location.
var related = sync.OnceValue(func() map[string][]string {
	all, _ := meta.All() // embedded catalogues: decoding is tested in meta
	out := map[string][]string{}
	for _, r := range all {
		for _, id := range r.SuppressedBy {
			out[r.ID] = append(out[r.ID], id)
			out[id] = append(out[id], r.ID)
		}
	}
	return out
})

// matcher consumes baseline entries for one filtering pass, reading the
// loaded counts lazily so that a pass over few findings copies nothing.
type matcher struct {
	b    *Baseline
	left map[key]int
}

func (m *matcher) take(k key) bool {
	n, seen := m.left[k]
	if !seen {
		n = m.b.counts[k]
	}
	if n == 0 {
		return false
	}
	m.left[k] = n - 1
	return true
}

// takeMoved consumes an entry of a rule related to k's at k's location,
// whatever its message (rules word their findings differently).
func (m *matcher) takeMoved(k key) bool {
	for _, other := range related()[k.rule] {
		for _, e := range m.b.byLoc[loc{k.path, other, k.line}] {
			if m.take(e) {
				return true
			}
		}
	}
	return false
}

// filter returns which of n findings, keyed by keyOf, the baseline
// covers. Exact entries are consumed first over all findings, so that a
// moved finding never takes the entry of an unmoved one.
func (b *Baseline) filter(n int, keyOf func(int) key) []bool {
	m := &matcher{b: b, left: map[key]int{}}
	covered := make([]bool, n)
	keys := make([]key, n)
	for i := range n {
		keys[i] = keyOf(i)
		covered[i] = m.take(keys[i])
	}
	for i := range n {
		if !covered[i] {
			covered[i] = m.takeMoved(keys[i])
		}
	}
	return covered
}

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
	if !filepath.IsAbs(dir) && !filepath.IsAbs(path) {
		if r, err := filepath.Rel(dir, path); err == nil {
			return filepath.ToSlash(r)
		}
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.ToSlash(path)
	}
	if r, err := filepath.Rel(dir, abs); err == nil {
		return filepath.ToSlash(r)
	}
	return filepath.ToSlash(abs)
}

// Recorder returns, for the file at path whose source is src, whether
// the baseline has an entry for one of its findings f: one of the
// finding's rule and message, or of a related rule (see related), at its
// line text, taken as FilterFindings takes it. It returns nil when the
// file has no entry. Unlike Filter, it consumes no entry, so a check can
// ask about a finding before the baseline is applied (see
// analysis.Engine.WithRecorded): earlier lists the findings reported
// before f with its rule and message, and those at the same line text take
// the entry's count first, as FilterFindings consumes it in source order.
// An entry of a related rule covers f whatever its count. The file's lines
// are indexed once, on the first question.
func (b *Baseline) Recorder(path string, src []byte) func(f diagnostic.Finding, earlier []diagnostic.Finding) bool {
	p := rel(b.dir, path)
	if !b.paths[p] {
		return nil
	}
	var text func(uint32) string
	return func(f diagnostic.Finding, earlier []diagnostic.Finding) bool {
		if text == nil {
			text = lineTexts(src)
		}
		t := text(f.Span.Start)
		if n := b.counts[key{p, f.Rule, f.Message, t}]; n > 0 {
			for _, e := range earlier {
				if text(e.Span.Start) == t {
					n--
				}
			}
			if n > 0 {
				return true
			}
		}
		for _, other := range related()[f.Rule] {
			if len(b.byLoc[loc{p, other, t}]) > 0 {
				return true
			}
		}
		return false
	}
}

// lineTexts indexes the lines of src once and returns the trimmed text of
// the line holding an offset: numbered by syntax.LineIndex (a lone CR ends
// a line) and split on newlines as Filter reads files (so a line after a
// lone CR keys as the text of the newline-split line of that number).
func lineTexts(src []byte) func(uint32) string {
	li := syntax.NewLineIndex(src)
	lines := strings.Split(string(bytes.ReplaceAll(src, []byte("\r\n"), []byte("\n"))), "\n")
	return func(offset uint32) string {
		if l, _ := li.Position(offset); l < len(lines) {
			return strings.TrimSpace(lines[l])
		}
		return ""
	}
}

// Filter removes items covered by the baseline and returns the rest plus
// the number of suppressed items.
func (b *Baseline) Filter(items []diagnostic.Item) ([]diagnostic.Item, int) {
	cache := map[string][]string{}
	covered := b.filter(len(items), func(i int) key {
		it := items[i]
		return key{rel(b.dir, it.Path), it.Rule, it.Message, lineText(cache, it.Path, it.Line)}
	})
	var out []diagnostic.Item
	suppressed := 0
	for i, it := range items {
		if covered[i] {
			suppressed++
			continue
		}
		out = append(out, it)
	}
	return out, suppressed
}

// FilterFindings removes one file's findings covered by the baseline,
// taking flagged lines from src (an editor buffer, say) instead of reading
// the file again. Lines are numbered and split exactly as Filter sees them.
func (b *Baseline) FilterFindings(path string, src []byte, findings []diagnostic.Finding) []diagnostic.Finding {
	covered := b.Covered(path, src, findings)
	if covered == nil {
		return findings
	}
	out := findings[:0:0]
	for i, f := range findings {
		if !covered[i] {
			out = append(out, f)
		}
	}
	return out
}

// Covered tells which of one file's findings the baseline covers, as
// FilterFindings filters them; nil when it covers none for lack of
// entries of the file.
func (b *Baseline) Covered(path string, src []byte, findings []diagnostic.Finding) []bool {
	p := rel(b.dir, path)
	// Most files have no entry: skip indexing their lines on every analysis.
	if len(findings) == 0 || !b.paths[p] {
		return nil
	}
	text := lineTexts(src)
	return b.filter(len(findings), func(i int) key {
		f := findings[i]
		return key{p, f.Rule, f.Message, text(f.Span.Start)}
	})
}

// Write creates a baseline at path covering items.
func Write(path string, items []diagnostic.Item) (int, error) {
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
