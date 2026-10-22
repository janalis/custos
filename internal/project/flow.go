package project

import (
	"context"
	"maps"
	"runtime"
	"slices"
	"sync"

	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

// BuildFlowSnapshot builds project-local summaries without including vendor
// bodies. Parsing is bounded by the same source reader as ordinary analysis.
// Trees are retained only during bounded composition, then released. Once
// ctx is done the build stops between files and rounds, and the snapshot is
// incomplete: callers check ctx before using it.
func BuildFlowSnapshot(ctx context.Context, files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options flow.Options) *flow.Snapshot {
	return BuildFlow(ctx, files, srcs, opt, ix, options).Snapshot
}

// FlowBuild is a flow snapshot with the summaries of each file it was
// composed from, so that Update redoes only what an edit invalidates.
type FlowBuild struct {
	Snapshot *flow.Snapshot
	files    []string
	at       map[string]int
	results  []*flow.File
	// sources are the supplied (unsaved) sources by path; the other files
	// are read from disk when summarized again.
	sources map[string][]byte
	opt     syntax.Options
	options flow.Options
}

// Files returns the files the build summarized, in build order.
func (b *FlowBuild) Files() []string { return b.files }

// BuildFlow is BuildFlowSnapshot keeping what Update needs.
func BuildFlow(ctx context.Context, files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options flow.Options) *FlowBuild {
	trees := make([]*syntax.File, len(files))
	// A file that crashes the parser keeps no tree: it is not summarized.
	_ = parallel(ctx, len(files), 0, func(i int) {
		var src []byte
		if i < len(srcs) {
			src = srcs[i]
		}
		trees[i] = parseFlowSource(files[i], src, opt)
	}, nil)
	b := &FlowBuild{files: files, at: make(map[string]int, len(files)), results: make([]*flow.File, len(trees)), sources: map[string][]byte{}, opt: opt, options: options}
	for i, path := range files {
		b.at[path] = i
		if i < len(srcs) && srcs[i] != nil {
			b.sources[path] = srcs[i]
		}
	}
	b.Snapshot = b.compose(ctx, flow.NewSnapshotWithOptions(options), make([]*flow.Snapshot, len(files)), ix, func(i int) *syntax.File { return trees[i] })
	return b
}

// Update returns the build for the same files after the sources of the
// changed ones (path → source, now supplied) changed, without changing any
// declaration other files see: the project index is the one the build
// used. The changed files are summarized again first; then, as in
// BuildFlow, every file whose summaries read a contract that changed, from
// its supplied source or else from disk. Files are parsed only when
// summarized again. ok is false when a changed path is not one of the
// build's files.
func (b *FlowBuild) Update(changed map[string][]byte, ix *index.Index) (*FlowBuild, bool) {
	dirty := make([]int, 0, len(changed))
	for path := range changed {
		i, ok := b.at[path]
		if !ok {
			return nil, false
		}
		dirty = append(dirty, i)
	}
	next := &FlowBuild{files: b.files, at: b.at, results: append([]*flow.File(nil), b.results...), sources: maps.Clone(b.sources), opt: b.opt, options: b.options}
	maps.Copy(next.sources, changed)
	var mu sync.Mutex
	trees := map[int]*syntax.File{}
	tree := func(i int) *syntax.File {
		mu.Lock()
		t, ok := trees[i]
		mu.Unlock()
		if !ok {
			t = parseFlowSource(b.files[i], next.sources[b.files[i]], b.opt)
			mu.Lock()
			trees[i] = t
			mu.Unlock()
		}
		return t
	}
	// The changed files are summarized once against the build's snapshot:
	// a file in no recursion keeps that summary, which its unchanged
	// contracts determine.
	probe := make([]*flow.File, len(dirty))
	_ = parallel(context.Background(), len(dirty), runtime.GOMAXPROCS(0)/2, func(j int) {
		if file := tree(dirty[j]); file != nil {
			probe[j] = extractFlow(file, ix, b.opt.Version, b.Snapshot, b.options)
		}
	}, nil)
	for j, i := range dirty {
		next.results[i] = probe[j]
	}
	// Within a recursion, the fixed point depends on where it starts: when
	// one depends on a changed file (reading, maybe through others, a
	// contract it declares itself), the files are summarized again from
	// none, as a full build does. Otherwise every file is summarized again
	// when a contract it read differs from the snapshot it was computed
	// against, as each summary follows from the contracts it read.
	if next.recursive(dirty) {
		srcs := make([][]byte, len(b.files))
		for i, path := range b.files {
			srcs[i] = next.sources[path]
		}
		return BuildFlow(context.Background(), b.files, srcs, b.opt, ix, b.options), true
	}
	against := make([]*flow.Snapshot, len(b.files))
	for i := range against {
		against[i] = b.Snapshot
	}
	next.Snapshot = next.compose(context.Background(), flow.NewSnapshotWithOptions(b.options, next.results...), against, ix, tree)
	return next, true
}

// recursive reports whether a file of seed, or one whose summaries read,
// directly or through others, a contract a file of seed declares, lies on
// a cycle of the reads relation (possibly reading its own contracts).
func (b *FlowBuild) recursive(seed []int) bool {
	owners := map[string][]int{}
	for j, r := range b.results {
		if r != nil {
			for key := range r.Summaries {
				owners[key] = append(owners[key], j)
			}
		}
	}
	// reads[i]: the files whose contracts the summaries of file i read.
	reads := make([][]int, len(b.results))
	readers := make([][]int, len(b.results))
	for i, r := range b.results {
		if r == nil {
			continue
		}
		for _, key := range r.Reads() {
			for _, j := range owners[key] {
				reads[i] = append(reads[i], j)
				readers[j] = append(readers[j], i)
			}
		}
	}
	affected := make([]bool, len(b.results))
	stack := append([]int(nil), seed...)
	for _, i := range seed {
		affected[i] = true
	}
	for len(stack) > 0 {
		i := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, r := range readers[i] {
			if !affected[r] {
				affected[r] = true
				stack = append(stack, r)
			}
		}
	}
	return slices.Contains(cyclicFiles(reads, affected), true)
}

// cyclicFiles marks the nodes among in that lie on a cycle of edges
// (strongly connected components with more than one node, or a node with
// an edge to itself), restricted to the nodes in.
func cyclicFiles(edges [][]int, in []bool) []bool {
	n := len(edges)
	index, low := make([]int, n), make([]int, n)
	for i := range index {
		index[i] = -1
	}
	onStack := make([]bool, n)
	cyclic := make([]bool, n)
	var stack []int
	next := 0
	var connect func(v int)
	connect = func(v int) {
		index[v], low[v] = next, next
		next++
		stack = append(stack, v)
		onStack[v] = true
		for _, w := range edges[v] {
			if !in[w] {
				continue
			}
			if w == v {
				cyclic[v] = true
			}
			if index[w] < 0 {
				connect(w)
				low[v] = min(low[v], low[w])
			} else if onStack[w] {
				low[v] = min(low[v], index[w])
			}
		}
		if low[v] == index[v] {
			var component []int
			for {
				w := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				onStack[w] = false
				component = append(component, w)
				if w == v {
					break
				}
			}
			if len(component) > 1 {
				for _, w := range component {
					cyclic[w] = true
				}
			}
		}
	}
	for v := range n {
		if in[v] && index[v] < 0 {
			connect(v)
		}
	}
	return cyclic
}

// parseFlowSource parses src, or the file at path when src is nil; nil
// when it cannot be read.
func parseFlowSource(path string, src []byte, opt syntax.Options) *syntax.File {
	if src == nil {
		var err error
		if src, err = ReadSource(path); err != nil {
			return nil
		}
	}
	return syntax.ParseBest(path, src, opt)
}

// maxRounds bounds the rounds of compose. A round redoes only the files
// whose summaries read a contract the previous one changed, and each round
// settles one more file of a chain of calls across files.
const maxRounds = 32

// maxSummaries bounds how often compose summarizes one file. A file of a
// chain is summarized once with its callees unknown and again once they
// settled, and one recursive through its own contracts needs a few rounds
// more; one still changing after that keeps oscillating, and summarizing it
// again only costs a round each time.
const maxSummaries = 8

// compose runs the bounded fixed point from snapshot: a file is summarized
// (again) when it has no summary or a contract its summaries read differs
// from the snapshot they were computed against (against). The files of a
// round read the snapshot of the round before: they are summarized side by
// side. What has not settled when it stops is dropped (see unsettled).
func (b *FlowBuild) compose(ctx context.Context, snapshot *flow.Snapshot, against []*flow.Snapshot, ix *index.Index, tree func(int) *syntax.File) *flow.Snapshot {
	summarized := make([]int, len(b.results))
	for range maxRounds {
		if ctx.Err() != nil {
			return snapshot
		}
		var work []int
		for i, r := range b.results {
			if (r == nil || !r.Current(against[i], snapshot)) && summarized[i] < maxSummaries {
				work = append(work, i)
				summarized[i]++
			}
		}
		// Half the cores: summaries allocate heavily, and on symfony/src
		// (16 cores) every core made the whole run about 8% slower.
		_ = parallel(ctx, len(work), runtime.GOMAXPROCS(0)/2, func(j int) {
			if file := tree(work[j]); file != nil {
				b.results[work[j]] = extractFlow(file, ix, b.opt.Version, snapshot, b.options)
				against[work[j]] = snapshot
			}
		}, nil)
		next := flow.NewSnapshotWithOptions(b.options, b.results...)
		if snapshot.Equal(next) {
			snapshot = next
			break
		}
		snapshot = next
	}
	return b.unsettled(snapshot, against)
}

// unsettled drops, once compose stops, every summary computed against
// contracts that differ from snapshot's, then those computed from dropped
// ones: they may claim what no settled contract supports. Calls into the
// files dropped are then unresolved, which their analyses report.
func (b *FlowBuild) unsettled(snapshot *flow.Snapshot, against []*flow.Snapshot) *flow.Snapshot {
	for {
		dropped := false
		for i, r := range b.results {
			if r != nil && !r.Current(against[i], snapshot) {
				b.results[i] = nil
				dropped = true
			}
		}
		if !dropped {
			return snapshot
		}
		snapshot = flow.NewSnapshotWithOptions(b.options, b.results...)
	}
}

// extractFlow is flow.Extract; tests replace it to simulate a crash. A file
// whose summaries crash keeps none: calls into it read as unresolved.
var extractFlow = flow.Extract
