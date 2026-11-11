package project

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"hash/maphash"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"custos/internal/platform/safeio"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

// The flow steps cache keeps, between runs, every summary the last flow
// build made of each analysed file with the contracts it read (the steps
// BuildFlowRecorded records), so that a run after an edit summarizes again
// only what the edit reaches: a file whose source is unchanged takes, as an
// update in the language server does, a recorded summary whose contracts
// still hold. Its results are those of a build without the cache.
//
// A summary depends on more than the contracts it read: on the project
// index, the dependency files the vendor provider summarizes, the options
// and the custos binary. The steps are therefore taken only while these are
// the same as when they were saved:
//
//   - the binary, the parse, inference and flow options, the working
//     directory (summaries hold paths as the run names them) and the vendor
//     directory are part of the key;
//   - so is the size and modification time of every file indexed but not
//     analysed, the trust the symbol cache gives its entries: a run in which
//     one was modified within racyWindow neither takes nor saves steps;
//   - the analysed files must be the same, each declaring what it declared
//     (index.Fingerprint, which ignores positions as an update does).
//
// Otherwise the build starts from none and saves its own steps.

// flowCacheMagic starts a flow steps file; a new layout changes it.
const flowCacheMagic = "custos flow steps 1\n"

// maxFlowCacheBytes bounds the flow steps file read back.
const maxFlowCacheBytes = 1 << 30

// flowSteps is what the flow steps cache holds for the analysed files of a
// run, and what the run needs to save its own.
type flowSteps struct {
	path, key string
	files     []string
	// sums holds the SHA-256 of each file's source, zero when unread;
	// prints the fingerprint of its declarations.
	sums, prints [][sha256.Size]byte
	// earlier holds the steps saved for the files whose source is
	// unchanged, nil when none can be taken.
	earlier *FlowBuild
	// current is set when the saved steps are those of these very sources:
	// a build that summarizes nothing anew leaves them as they are.
	current bool
}

// flowStepsPath is where the cache keeps the flow steps.
func (c *SymbolCache) flowStepsPath() string {
	return strings.TrimSuffix(c.path, ".gob") + "-flow.bin"
}

// flowKey is the key of the flow steps for options; ok is false when a
// file indexed but not analysed was modified within racyWindow: its stat
// does not tell whether it changes again.
func (c *SymbolCache) flowKey(root string, options flow.Options) (key string, ok bool) {
	if c.racy {
		return "", false
	}
	cwd, _ := os.Getwd()
	h := sha256.New()
	for _, path := range slices.Sorted(maps.Keys(c.fresh)) {
		e := c.fresh[path]
		fmt.Fprintf(h, "%s\x00%d\x00%d\n", path, e.Size, e.ModTime)
	}
	vendor := absolute(filepath.Join(root, vendorDir(root)))
	return fmt.Sprintf("%s\x00%s\x00%s\x00%s\x00%x", c.key, options.String(), cwd, vendor, h.Sum(nil)), true
}

// fingerprint makes the index record the fingerprint of the declarations of
// the first n files it indexes (the analysed ones), for the flow steps.
func (c *SymbolCache) fingerprint(n int) {
	if c != nil {
		c.prints = make([][sha256.Size]byte, n)
	}
}

// declared records the declarations of the analysed file i (see
// fingerprint), before DropStaleInferred rewrites them.
func (c *SymbolCache) declared(i int, fs *index.FileSymbols) {
	if c != nil && c.prints != nil {
		c.prints[i] = index.Fingerprint(fs)
	}
}

// flowSteps reads the steps saved for files (srcs holds their sources, nil
// where unread), keyed as flowKey says; the result is nil when the run can
// neither take nor save steps.
func (c *SymbolCache) flowSteps(root string, files []string, srcs [][]byte, options flow.Options) *flowSteps {
	if c == nil || c.prints == nil {
		return nil
	}
	key, ok := c.flowKey(root, options)
	if !ok {
		return nil
	}
	s := &flowSteps{path: c.flowStepsPath(), key: key, files: files, sums: make([][sha256.Size]byte, len(files)), prints: c.prints}
	_ = parallel(context.Background(), len(files), 0, func(i int) {
		if srcs[i] != nil {
			s.sums[i] = sha256.Sum256(srcs[i])
		}
	}, nil)
	b, err := safeio.ReadFile(s.path, maxFlowCacheBytes)
	if err != nil {
		return s
	}
	s.earlier, s.current = s.decode(b, srcs, options)
	return s
}

// decode returns, from the saved steps b, the earlier build whose steps the
// files with an unchanged source take, and whether every source is
// unchanged; nil when b does not decode, is keyed otherwise or the index
// changed.
func (s *flowSteps) decode(b []byte, srcs [][]byte, options flow.Options) (*FlowBuild, bool) {
	rest, ok := bytes.CutPrefix(b, []byte(flowCacheMagic))
	if !ok {
		return nil, false
	}
	size, n := binary.Uvarint(rest)
	if n <= 0 || size > uint64(len(rest)-n) || string(rest[n:n+int(size)]) != s.key {
		return nil, false
	}
	d, err := flow.NewDecoder(rest[n+int(size):], options)
	if err != nil {
		return nil, false
	}
	r := d.Reader()
	count, isNil := r.Len()
	if isNil || count != len(s.files) {
		return nil, false
	}
	earlier := &FlowBuild{files: s.files, at: make(map[string]int, len(s.files)), steps: make([][]flowStep, len(s.files)), hashes: make([]uint64, len(s.files))}
	current := true
	for i, path := range s.files {
		// The files are saved in the order the run lists them.
		if r.String() != path || r.String() != string(s.prints[i][:]) {
			return nil, false
		}
		same := r.String() == string(s.sums[i][:]) && srcs[i] != nil
		current = current && same
		n, _ := r.Len()
		steps := make([]flowStep, n)
		for k := range steps {
			steps[k] = flowStep{d.File(), d.Snapshot()}
		}
		if same && len(steps) > 0 {
			earlier.at[path] = i
			earlier.steps[i] = steps
			earlier.hashes[i] = maphash.Bytes(sourceSeed, srcs[i])
		}
	}
	if !r.Done() {
		return nil, false
	}
	return earlier, current
}

// save writes the steps build recorded, unless they are those read back.
// A file whose steps cannot be encoded is saved without them: it is
// summarized anew next time.
func (s *flowSteps) save(build *FlowBuild) error {
	if s == nil || s.current && build.made.Load() == 0 {
		return nil
	}
	e := flow.NewEncoder()
	w := e.Writer()
	w.Len(len(s.files), false)
	for i, path := range s.files {
		w.String(path)
		w.String(string(s.prints[i][:]))
		w.String(string(s.sums[i][:]))
		mark := e.Mark()
		steps := build.steps[i]
		w.Len(len(steps), false)
		for _, step := range steps {
			e.File(step.result)
			e.Snapshot(step.against)
		}
		if e.Err() != nil {
			e.Rollback(mark)
			w.Len(0, false)
		}
	}
	out := binary.AppendUvarint([]byte(flowCacheMagic), uint64(len(s.key)))
	out = append(append(out, s.key...), e.Bytes()...)
	return writeCacheBytes(s.path, out)
}
