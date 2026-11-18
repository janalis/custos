package project

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"
	"sync/atomic"

	"custos/internal/diagnostic"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/platform/safeio"
	"custos/internal/semantic/flow"
)

// The findings cache keeps, between runs of AnalyzeReport, what the
// analysis of each file reported, so that a run takes it back for a file
// nothing it may depend on changed in, instead of analyzing it again. The
// results are those of a run without the cache.
//
// A file's findings follow from its source, the project index, the flow
// snapshot, the engine and what the checks see of the baseline. An entry is
// therefore taken only while
//
//   - the file's source has the same SHA-256;
//   - the run key is the same: the binary and the parse and inference
//     options (the symbol cache's key), the working directory, the project
//     root, each file's path within its project (see
//     analysis.Config.ProjectPath), the enabled rules and their
//     configuration, --parse-once, the
//     baseline the checks see (Options.RecordedKey), the Composer world's
//     inputs, the size and modification time of every file indexed but not
//     analyzed, and the declarations of every analyzed file
//     (index.Fingerprint, as for the flow steps);
//   - the flow snapshot the rules read equals the one the entries were
//     computed against.
//
// A run in which a file indexed but not analyzed was modified within
// racyWindow, or whose engine sees a baseline it was not told the key of,
// neither takes nor saves entries; a file that could not be read is never
// kept.

// findingsCacheMagic starts a findings file; a new layout changes it.
const findingsCacheMagic = "custos findings 1\n"

// maxFindingsCacheBytes bounds the findings file read back.
const maxFindingsCacheBytes = 1 << 30

type findingsCache struct {
	path, key string
	snapshot  *flow.Snapshot
	options   flow.Options
	// old holds the entries read back, by path; nil when none can be taken.
	old map[string]cachedResult
	// fresh holds, by file index, the entry of each file of this run; kept
	// counts those taken back.
	fresh []*cachedResult
	files []string
	kept  atomic.Int64
}

type cachedResult struct {
	Sum        [sha256.Size]byte
	Findings   []cachedFinding
	Incomplete []diagnostic.AnalysisIssue
	Errors     []syntax.Error
}

type cachedFinding struct {
	Rule     string
	Severity diagnostic.Severity
	Span     syntax.Span
	Message  string
	// Fixes holds the titles of the finding's fixes.
	Fixes []string
}

// findingsInputs is what the run key covers besides the symbol cache's.
type findingsInputs struct {
	root, recorded string
	parseOnce      bool
	engine         *analysis.Engine
	// world digests the inputs of the Composer world.
	world string
	// php lists the analyzed PHP files, whose declarations c.prints holds.
	php []string
}

// findingsPath is where the cache keeps the findings.
func (c *SymbolCache) findingsPath() string {
	return strings.TrimSuffix(c.path, ".gob") + "-findings.bin"
}

// findings opens the findings cache of a run analyzing files against
// snapshot (nil without flow); nil when the run can neither take nor save
// findings.
func (c *SymbolCache) findings(files []string, in findingsInputs, snapshot *flow.Snapshot, options flow.Options) *findingsCache {
	if c == nil || c.racy || c.prints == nil || in.engine.Recording() && in.recorded == "" {
		return nil
	}
	cfg := in.engine.Config()
	engine, _ := json.Marshal(struct {
		Rules    []string
		PHP      int
		Style    analysis.ComparisonStyle
		Settings map[string]analysis.RuleConfig
		Flow     string
	}{in.engine.Rules(), int(cfg.PHP), cfg.ComparisonStyle, cfg.Rules, cfg.Flow.String()})
	cwd, _ := os.Getwd()
	h := sha256.New()
	fmt.Fprintf(h, "%s\x00%s\x00%s\x00%s\x00%t\x00%s\x00%s\n", c.key, cwd, absolute(in.root), engine, in.parseOnce, in.recorded, in.world)
	for _, path := range slices.Sorted(maps.Keys(c.fresh)) {
		e := c.fresh[path]
		fmt.Fprintf(h, "%s\x00%d\x00%d\n", path, e.Size, e.ModTime)
	}
	for i, path := range in.php {
		fmt.Fprintf(h, "%s\x00%x\n", path, c.prints[i])
	}
	// Rules matching path globs see each file's path within its project,
	// which a custos.json or composer.json anywhere above it decides.
	if relate := cfg.ProjectPath; relate != nil {
		for _, path := range files {
			fmt.Fprintf(h, "%s\x00%s\n", path, relate(path))
		}
	}
	fc := &findingsCache{path: c.findingsPath(), key: fmt.Sprintf("%x", h.Sum(nil)), snapshot: snapshot, options: options, fresh: make([]*cachedResult, len(files)), files: files}
	if b, err := safeio.ReadFile(fc.path, maxFindingsCacheBytes); err == nil {
		fc.old = fc.decode(b)
	}
	return fc
}

// decode returns the entries b holds, nil unless it is keyed as fc and was
// computed against an equal snapshot.
func (fc *findingsCache) decode(b []byte) map[string]cachedResult {
	key, rest, ok := cutField(b, findingsCacheMagic)
	if !ok || string(key) != fc.key {
		return nil
	}
	snap, rest, ok := cutField(rest, "")
	if !ok {
		return nil
	}
	d, err := flow.NewDecoder(snap, fc.options)
	if err != nil {
		return nil
	}
	saved := d.Snapshot()
	if !d.Reader().Done() || (saved == nil) != (fc.snapshot == nil) || saved != nil && !saved.Equal(fc.snapshot) {
		return nil
	}
	var entries map[string]cachedResult
	if gob.NewDecoder(bytes.NewReader(rest)).Decode(&entries) != nil {
		return nil
	}
	return entries
}

// worldDigest digests what ComposerWorld reads besides the index: the
// Composer metadata under root and what the walks listed and left out.
func worldDigest(root string, indexed []string, skipped Skipped) string {
	h := sha256.New()
	fmt.Fprintf(h, "%s\n", ComposerWorldKey(root))
	_, installed := WorldInputs(root)
	if b, err := safeio.ReadFile(installed, maxFindingsCacheBytes); err == nil {
		fmt.Fprintf(h, "%x\n", sha256.Sum256(b))
	}
	for _, list := range [][]string{indexed, skipped.Dirs, skipped.Files} {
		fmt.Fprintf(h, "%d\x00%s\n", len(list), strings.Join(slices.Sorted(slices.Values(list)), "\x00"))
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// cutField cuts prefix, then a length-prefixed field, from b.
func cutField(b []byte, prefix string) (field, rest []byte, ok bool) {
	if b, ok = bytes.CutPrefix(b, []byte(prefix)); !ok {
		return nil, nil, false
	}
	size, n := binary.Uvarint(b)
	if n <= 0 || size > uint64(len(b)-n) {
		return nil, nil, false
	}
	return b[n : n+int(size)], b[n+int(size):], true
}

// take returns the result kept for file i, read as src, when its entry
// holds; otherwise it remembers the source's sum for store.
func (fc *findingsCache) take(i int, src []byte) (diagnostic.FileResult, bool) {
	if fc == nil {
		return diagnostic.FileResult{}, false
	}
	sum := sha256.Sum256(src)
	e, ok := fc.old[fc.files[i]]
	if !ok || e.Sum != sum {
		fc.fresh[i] = &cachedResult{Sum: sum}
		return diagnostic.FileResult{}, false
	}
	fc.fresh[i] = &e
	fc.kept.Add(1)
	r := diagnostic.FileResult{Path: fc.files[i], Src: src, Errors: e.Errors, Incomplete: e.Incomplete}
	for _, f := range e.Findings {
		finding := diagnostic.Finding{Rule: f.Rule, Severity: f.Severity, Span: f.Span, Message: f.Message}
		for _, title := range f.Fixes {
			finding.Fixes = append(finding.Fixes, diagnostic.Fix{Title: title})
		}
		r.Findings = append(r.Findings, finding)
	}
	return r, true
}

// store keeps result, the analysis of file i after take missed.
func (fc *findingsCache) store(i int, result diagnostic.FileResult) {
	if fc == nil {
		return
	}
	e := fc.fresh[i]
	e.Errors, e.Incomplete = result.Errors, result.Incomplete
	for _, f := range result.Findings {
		cf := cachedFinding{Rule: f.Rule, Severity: f.Severity, Span: f.Span, Message: f.Message}
		for _, fix := range f.Fixes {
			cf.Fixes = append(cf.Fixes, fix.Title)
		}
		e.Findings = append(e.Findings, cf)
	}
}

// save writes the entries of this run, unless they are those read back.
func (fc *findingsCache) save() error {
	if fc == nil {
		return nil
	}
	entries := make(map[string]cachedResult, len(fc.fresh))
	unchanged := len(fc.old) == len(fc.files)
	for i, e := range fc.fresh {
		if e == nil {
			continue // unread
		}
		entries[fc.files[i]] = *e
		old, ok := fc.old[fc.files[i]]
		unchanged = unchanged && ok && old.Sum == e.Sum
	}
	if unchanged && len(entries) == len(fc.old) {
		return nil
	}
	enc := flow.NewEncoder()
	enc.Snapshot(fc.snapshot.Contracts())
	if enc.Err() != nil {
		return enc.Err()
	}
	var body bytes.Buffer
	// Encoding plain exported data into memory cannot fail.
	_ = gob.NewEncoder(&body).Encode(entries)
	snap := enc.Bytes()
	out := binary.AppendUvarint([]byte(findingsCacheMagic), uint64(len(fc.key)))
	out = append(out, fc.key...)
	out = binary.AppendUvarint(out, uint64(len(snap)))
	out = append(append(out, snap...), body.Bytes()...)
	return writeCacheBytes(fc.path, out)
}
