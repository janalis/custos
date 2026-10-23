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
	// order is the order the build took (see BuildFlowOrdered), nil for
	// none: a build of all files from none again takes it.
	order map[string]int
	// traced is set when seen and declares hold, for each file, the keys
	// of the contracts every summary of it consulted on the way to its
	// results, and those its summaries declare, as a build of all files
	// from none computes them (see rescope).
	traced   bool
	seen     []map[string]bool
	declares [][]string
	// stale marks the files whose seen is only a hint: an update outside
	// any recursion changed what they consult on the way (see retrace).
	stale []bool
}

// Files returns the files the build summarized, in build order.
func (b *FlowBuild) Files() []string { return b.files }

// BuildFlow is BuildFlowSnapshot keeping what Update needs.
func BuildFlow(ctx context.Context, files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options flow.Options) *FlowBuild {
	return BuildFlowOrdered(ctx, files, srcs, opt, ix, options, nil)
}

// maxWaves bounds the waves of BuildFlowOrdered: each one rebuilds the
// snapshot and waits for its slowest file, and the deeper files are few.
const maxWaves = 16

// BuildFlowOrdered is BuildFlow summarizing each file, the first time, after
// the files whose contracts it read in an earlier build: order gives their
// depth by path (see Order), and a path it lacks has depth 0. Files are
// summarized in waves of growing depth, the deepest together, each wave
// against the snapshot of the waves before; composition then goes on as in
// BuildFlow. Most files thus read settled contracts the first time instead
// of being summarized again once they settle. Where BuildFlow settles
// every file, the snapshot is the same; a call chain whose end BuildFlow
// drops past its rounds may settle here instead.
func BuildFlowOrdered(ctx context.Context, files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options flow.Options, order map[string]int) *FlowBuild {
	trees := make([]*syntax.File, len(files))
	// A file that crashes the parser keeps no tree: it is not summarized.
	_ = parallel(ctx, len(files), 0, func(i int) {
		var src []byte
		if i < len(srcs) {
			src = srcs[i]
		}
		trees[i] = parseFlowSource(files[i], src, opt)
	}, nil)
	b := &FlowBuild{files: files, at: make(map[string]int, len(files)), results: make([]*flow.File, len(trees)), sources: map[string][]byte{}, opt: opt, options: options, order: order, traced: true, seen: make([]map[string]bool, len(files)), declares: make([][]string, len(files))}
	for i, path := range files {
		b.at[path] = i
		if i < len(srcs) && srcs[i] != nil {
			b.sources[path] = srcs[i]
		}
	}
	tree := func(i int) *syntax.File { return trees[i] }
	snapshot := flow.NewSnapshotWithOptions(options)
	against := make([]*flow.Snapshot, len(files))
	if len(order) > 0 {
		waves := make([][]int, maxWaves)
		for i, path := range files {
			depth := min(order[path], maxWaves-1)
			waves[depth] = append(waves[depth], i)
		}
		for _, wave := range waves {
			if len(wave) > 0 {
				b.summarize(ctx, wave, snapshot, against, ix, tree)
				snapshot = flow.NewSnapshotWithOptions(options, b.results...)
			}
		}
	}
	b.Snapshot = b.compose(ctx, snapshot, against, ix, tree)
	return b
}

// Order returns, by path, the depth of each file in the relation of the
// contracts its summaries read: 0 for a file reading none of another file,
// else one more than the deepest file it reads, ignoring the reads that
// close a cycle. BuildFlowOrdered takes it for a later build.
func (b *FlowBuild) Order() map[string]int {
	order := make(map[string]int, len(b.files))
	for i, depth := range b.depths() {
		order[b.files[i]] = depth
	}
	return order
}

// depths is Order by file index.
func (b *FlowBuild) depths() []int {
	reads := b.reads()
	depth := make([]int, len(b.files))
	// state: 0 unvisited, 1 on the path being walked, 2 done.
	state := make([]uint8, len(b.files))
	var walk func(i int)
	walk = func(i int) {
		state[i] = 1
		for _, j := range reads[i] {
			if state[j] == 0 {
				walk(j)
			}
			if state[j] == 2 {
				depth[i] = max(depth[i], depth[j]+1)
			}
		}
		state[i] = 2
	}
	for i := range b.files {
		if state[i] == 0 {
			walk(i)
		}
	}
	return depth
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
	next := &FlowBuild{files: b.files, at: b.at, results: append([]*flow.File(nil), b.results...), sources: maps.Clone(b.sources), opt: b.opt, options: b.options, order: b.order}
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
	// none, as a full build does (in the order b took, if any): all of them
	// unless rescope can tell which. Otherwise every file is
	// summarized again when a contract it read differs from the snapshot it
	// was computed against, as each summary follows from the contracts it
	// read; retrace tells which records that leaves stale.
	if next.recursive(dirty) {
		if scoped := b.rescope(next.sources, dirty, probe, ix); scoped != nil {
			return scoped, true
		}
		srcs := make([][]byte, len(b.files))
		for i, path := range b.files {
			srcs[i] = next.sources[path]
		}
		return BuildFlowOrdered(context.Background(), b.files, srcs, b.opt, ix, b.options, b.order), true
	}
	against := make([]*flow.Snapshot, len(b.files))
	for i := range against {
		against[i] = b.Snapshot
	}
	next.Snapshot = next.compose(context.Background(), flow.NewSnapshotWithOptions(b.options, next.results...), against, ix, tree)
	b.retrace(next, dirty, probe)
	return next, true
}

// retrace carries the record of b over to next, which Update made from b
// summarizing again against b's snapshot after the files of dirty changed
// (probe holds their first summaries). A file that consulted, maybe through
// others, a contract of a changed file may consult others on the way in a
// build of all files from none: its record becomes stale, a hint that
// rescope summarizes again. The others stand.
func (b *FlowBuild) retrace(next *FlowBuild, dirty []int, probe []*flow.File) {
	if !b.traced {
		return
	}
	declares := slices.Clone(b.declares)
	seed := slices.Clone(dirty)
	for j, i := range dirty {
		if probe[j] == nil {
			return // what it declares is unknown
		}
		declares[i] = union(b.declares[i], slices.Collect(maps.Keys(probe[j].Summaries)))
	}
	for i, stale := range b.stale {
		if stale {
			seed = append(seed, i)
		}
	}
	_, readers := b.links(keyOwners(declares))
	seen := slices.Clone(b.seen)
	for i, r := range next.results {
		if r != nil && r != b.results[i] {
			hint := maps.Clone(seen[i])
			if hint == nil {
				hint = map[string]bool{}
			}
			for _, key := range r.Consulted() {
				hint[key] = true
			}
			seen[i] = hint
		}
	}
	next.traced, next.seen, next.declares, next.stale = true, seen, declares, closure(seed, readers)
}

// rescope returns the build of all files from none after the files of dirty
// changed (sources as Update has them; probe holds their summaries), made
// by summarizing from none only the files that build may summarize
// differently and those whose contracts they consult; nil when that cannot
// be told. A file outside them consulted, every time it was summarized, no
// contract of a file the change reaches, and those inside consult only
// contracts declared inside, so each is summarized exactly as the build of
// all files summarizes it.
func (b *FlowBuild) rescope(sources map[string][]byte, dirty []int, probe []*flow.File, ix *index.Index) *FlowBuild {
	if !b.traced {
		return nil
	}
	declares := slices.Clone(b.declares)
	for j, i := range dirty {
		if probe[j] == nil {
			return nil
		}
		// The keys the file declared, and those it now declares.
		declares[i] = union(b.declares[i], slices.Collect(maps.Keys(probe[j].Summaries)))
	}
	owners := keyOwners(declares)
	reads, readers := b.links(owners)
	// Files with a stale record are summarized again as if the edit
	// reached them.
	seed := slices.Clone(dirty)
	for i, stale := range b.stale {
		if stale {
			seed = append(seed, i)
		}
	}
	reached := closure(seed, readers)
	seed = seed[:0]
	for i, in := range reached {
		if in {
			seed = append(seed, i)
		}
	}
	// The contracts the changed files consult now, as far as the probe
	// tells, are likely consulted on the way too.
	for _, r := range probe {
		for _, key := range r.Consulted() {
			seed = append(seed, owners[key]...)
		}
	}
	scope := closure(seed, reads)
	var files []string
	var srcs [][]byte
	var at []int
	for i, in := range scope {
		if in {
			files, srcs, at = append(files, b.files[i]), append(srcs, sources[b.files[i]]), append(at, i)
		}
	}
	if len(files) == len(b.files) {
		return nil
	}
	part := BuildFlowOrdered(context.Background(), files, srcs, b.opt, ix, b.options, b.order)
	for k, i := range at {
		declares[i] = part.declares[k]
	}
	owners = keyOwners(declares)
	for k, keys := range part.seen {
		for key := range keys {
			for _, j := range owners[key] {
				if !scope[j] {
					return nil // a contract declared outside: unlike the full build
				}
			}
		}
		if i := at[k]; !reached[i] && !sameSummaries(part.results[k], b.results[i]) {
			return nil
		}
	}
	next := &FlowBuild{files: b.files, at: b.at, results: slices.Clone(b.results), sources: sources, opt: b.opt, options: b.options, order: b.order, traced: true, seen: slices.Clone(b.seen), declares: declares}
	for k, i := range at {
		next.results[i], next.seen[i] = part.results[k], part.seen[k]
	}
	next.Snapshot = flow.NewSnapshotWithOptions(b.options, next.results...)
	return next
}

// links returns, for each file, the files declaring (by owners) a contract
// its record says it consulted, and the files whose record says they
// consulted one it declares.
func (b *FlowBuild) links(owners map[string][]int) (reads, readers [][]int) {
	reads = make([][]int, len(b.files))
	readers = make([][]int, len(b.files))
	for i, keys := range b.seen {
		for key := range keys {
			for _, j := range owners[key] {
				reads[i] = append(reads[i], j)
				readers[j] = append(readers[j], i)
			}
		}
	}
	return reads, readers
}

// union returns the keys of a and b, sorted, each once.
func union(a, b []string) []string {
	keys := slices.Concat(a, b)
	slices.Sort(keys)
	return slices.Compact(keys)
}

// keyOwners returns the files declaring each key.
func keyOwners(declares [][]string) map[string][]int {
	owners := map[string][]int{}
	for i, keys := range declares {
		for _, key := range keys {
			owners[key] = append(owners[key], i)
		}
	}
	return owners
}

// closure returns the nodes reached from seed along edges, seed included.
func closure(seed []int, edges [][]int) []bool {
	in := make([]bool, len(edges))
	stack := slices.Clone(seed)
	for _, i := range seed {
		in[i] = true
	}
	for len(stack) > 0 {
		i := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for _, j := range edges[i] {
			if !in[j] {
				in[j] = true
				stack = append(stack, j)
			}
		}
	}
	return in
}

// sameSummaries reports that a and b (either nil) hold the same summaries.
func sameSummaries(a, b *flow.File) bool {
	if a == nil || b == nil {
		return a == b
	}
	return flow.NewSnapshot(a).Equal(flow.NewSnapshot(b))
}

// recursive reports whether a file of seed, or one whose summaries read,
// directly or through others, a contract a file of seed declares, lies on
// a cycle of the reads relation (possibly reading its own contracts).
func (b *FlowBuild) recursive(seed []int) bool {
	reads := b.reads()
	readers := make([][]int, len(b.results))
	for i, files := range reads {
		for _, j := range files {
			readers[j] = append(readers[j], i)
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

// reads returns, for each file, the files whose contracts its summaries
// read.
func (b *FlowBuild) reads() [][]int {
	owners := map[string][]int{}
	for j, r := range b.results {
		if r != nil {
			for key := range r.Summaries {
				owners[key] = append(owners[key], j)
			}
		}
	}
	reads := make([][]int, len(b.results))
	for i, r := range b.results {
		if r != nil {
			for _, key := range r.Reads() {
				reads[i] = append(reads[i], owners[key]...)
			}
		}
	}
	return reads
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
		b.summarize(ctx, work, snapshot, against, ix, tree)
		next := flow.NewSnapshotWithOptions(b.options, b.results...)
		if snapshot.Equal(next) {
			snapshot = next
			break
		}
		snapshot = next
	}
	return b.unsettled(snapshot, against)
}

// summarize summarizes the files of work side by side against snapshot.
func (b *FlowBuild) summarize(ctx context.Context, work []int, snapshot *flow.Snapshot, against []*flow.Snapshot, ix *index.Index, tree func(int) *syntax.File) {
	// Half the cores: summaries allocate heavily, and on symfony/src
	// (16 cores) every core made the whole run about 8% slower.
	_ = parallel(ctx, len(work), runtime.GOMAXPROCS(0)/2, func(j int) {
		if file := tree(work[j]); file != nil {
			i := work[j]
			b.results[i] = extractFlow(file, ix, b.opt.Version, snapshot, b.options)
			against[i] = snapshot
			if b.traced && b.results[i] != nil {
				b.trace(i)
			}
		}
	}, nil)
}

// trace records what the summaries of file i just consulted and declare.
func (b *FlowBuild) trace(i int) {
	r := b.results[i]
	if b.seen[i] == nil {
		b.seen[i] = map[string]bool{}
		b.declares[i] = slices.Sorted(maps.Keys(r.Summaries))
	}
	for _, key := range r.Consulted() {
		b.seen[i][key] = true
	}
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
