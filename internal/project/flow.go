package project

import (
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
// Trees are retained only during bounded composition, then released.
func BuildFlowSnapshot(files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options ...flow.Options) *flow.Snapshot {
	return BuildFlow(files, srcs, opt, ix, options...).Snapshot
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
func BuildFlow(files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options ...flow.Options) *FlowBuild {
	var resolved flow.Options
	if len(options) > 0 {
		resolved = options[0]
	}
	trees := make([]*syntax.File, len(files))
	each(len(files), runtime.GOMAXPROCS(0), func(i int) {
		var src []byte
		if i < len(srcs) {
			src = srcs[i]
		}
		trees[i] = parseFlowSource(files[i], src, opt)
	})
	b := &FlowBuild{files: files, at: make(map[string]int, len(files)), results: make([]*flow.File, len(trees)), sources: map[string][]byte{}, opt: opt, options: resolved}
	for i, path := range files {
		b.at[path] = i
		if i < len(srcs) && srcs[i] != nil {
			b.sources[path] = srcs[i]
		}
	}
	b.Snapshot = b.compose(flow.NewSnapshotWithOptions(resolved), make([]*flow.Snapshot, len(files)), ix, func(i int) *syntax.File { return trees[i] })
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
	each(len(dirty), max(1, runtime.GOMAXPROCS(0)/2), func(j int) {
		if file := tree(dirty[j]); file != nil {
			probe[j] = flow.Extract(file, ix, b.opt.Version, b.Snapshot, b.options)
		}
	})
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
		return BuildFlow(b.files, srcs, b.opt, ix, b.options), true
	}
	against := make([]*flow.Snapshot, len(b.files))
	for i := range against {
		against[i] = b.Snapshot
	}
	next.Snapshot = next.compose(flow.NewSnapshotWithOptions(b.options, next.results...), against, ix, tree)
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

// compose runs the bounded fixed point from snapshot: a file is summarized
// (again) when it has no summary or a contract its summaries read differs
// from the snapshot they were computed against (against). The files of a
// round read the snapshot of the round before: they are summarized side by
// side.
func (b *FlowBuild) compose(snapshot *flow.Snapshot, against []*flow.Snapshot, ix *index.Index, tree func(int) *syntax.File) *flow.Snapshot {
	for round := 0; round < flow.MaxDepth; round++ {
		var work []int
		for i, r := range b.results {
			if r == nil || !r.Current(against[i], snapshot) {
				work = append(work, i)
			}
		}
		each(len(work), max(1, runtime.GOMAXPROCS(0)/2), func(j int) {
			if file := tree(work[j]); file != nil {
				b.results[work[j]] = flow.Extract(file, ix, b.opt.Version, snapshot, b.options)
				against[work[j]] = snapshot
			}
		})
		next := flow.NewSnapshotWithOptions(b.options, b.results...)
		if snapshot.Equal(next) {
			return next
		}
		snapshot = next
	}
	return snapshot
}

// each runs work for every index below n on that many goroutines and returns
// when all are done. Each call of work owns its index.
func each(n, workers int, work func(i int)) {
	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				work(i)
			}
		}()
	}
	for i := range n {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
}
