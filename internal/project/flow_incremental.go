package project

import (
	"context"
	"hash/maphash"
	"maps"
	"runtime"
	"slices"
	"sync"

	"custos/internal/php/syntax"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

// Incremental updates: a build summarizes again only what an edit reaches.

// Update returns the build for the same files after the sources of the
// changed ones (path → source, now supplied) changed, without changing any
// declaration other files see: the project index is the one the build
// used. The changed files are summarized again first; then, as in
// BuildFlow, every file whose summaries read a contract that changed, from
// its supplied source or else from disk. Files are parsed only when
// summarized again. ok is false when a changed path is not one of the
// build's files, or when ctx is done before the update completes (what it
// summarized is then dropped).
func (b *FlowBuild) Update(ctx context.Context, changed map[string][]byte, ix *index.Index) (*FlowBuild, bool) {
	dirty := make([]int, 0, len(changed))
	for path := range changed {
		i, ok := b.at[path]
		if !ok {
			return nil, false
		}
		dirty = append(dirty, i)
	}
	next := &FlowBuild{files: b.files, at: b.at, results: append([]*flow.File(nil), b.results...), sources: maps.Clone(b.sources), opt: b.opt, options: b.options, order: b.order, sizes: b.sizes}
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
	if parallel(ctx, len(dirty), summaryWorkers(runtime.GOMAXPROCS(0)), func(j int) {
		if file := tree(dirty[j]); file != nil {
			probe[j] = extractFlow(file, ix, b.opt.Version, b.Snapshot, b.options, nil, nil, nil)
		}
	}, nil) != nil {
		return nil, false
	}
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
		if scoped := b.rescope(ctx, next.sources, dirty, probe, ix); scoped != nil {
			return done(ctx, scoped)
		}
		srcs := make([][]byte, len(b.files))
		for i, path := range b.files {
			srcs[i] = next.sources[path]
		}
		return done(ctx, buildFlow(ctx, b.files, srcs, b.opt, ix, b.options, b.order, b, b.steps != nil, nil))
	}
	against := make([]*flow.Snapshot, len(b.files))
	for i := range against {
		against[i] = b.Snapshot
	}
	next.Snapshot = next.compose(ctx, flow.NewSnapshotWithOptions(b.options, next.results...), against, ix, tree)
	if ctx.Err() != nil {
		return nil, false
	}
	b.retrace(next, dirty, probe)
	b.restep(next, dirty, against)
	return next, true
}

// done returns build, unless ctx was done before it completed.
func done(ctx context.Context, build *FlowBuild) (*FlowBuild, bool) {
	if ctx.Err() != nil {
		return nil, false
	}
	return build, true
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

// restep carries the steps b recorded over to next, which Update made from
// b: a file summarized again holds its last summary alone, with the source
// it was made from.
func (b *FlowBuild) restep(next *FlowBuild, dirty []int, against []*flow.Snapshot) {
	if b.steps == nil {
		return
	}
	next.steps, next.hashes = slices.Clone(b.steps), slices.Clone(b.hashes)
	edited := make([]bool, len(b.files))
	for _, i := range dirty {
		next.hashes[i], edited[i] = maphash.Bytes(sourceSeed, next.sources[b.files[i]]), true
	}
	for i, r := range next.results {
		// The steps of an edited file describe its former source.
		if r != b.results[i] || edited[i] {
			next.steps[i] = nil
			if r != nil {
				next.steps[i] = []flowStep{{r, r.Basis(against[i])}}
			}
		}
	}
}

// rescope returns the build of all files from none after the files of dirty
// changed (sources as Update has them; probe holds their summaries), made
// by summarizing from none only the files that build may summarize
// differently and those whose contracts they consult; nil when that cannot
// be told. A file outside them consulted, every time it was summarized, no
// contract of a file the change reaches, and those inside consult only
// contracts declared inside, so each is summarized exactly as the build of
// all files summarizes it.
func (b *FlowBuild) rescope(ctx context.Context, sources map[string][]byte, dirty []int, probe []*flow.File, ix *index.Index) *FlowBuild {
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
	part := buildFlow(ctx, files, srcs, b.opt, ix, b.options, b.order, b, b.steps != nil, nil)
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
	next := &FlowBuild{files: b.files, at: b.at, results: slices.Clone(b.results), sources: sources, opt: b.opt, options: b.options, order: b.order, traced: true, seen: slices.Clone(b.seen), declares: declares, sizes: b.sizes}
	for k, i := range at {
		next.results[i], next.seen[i] = part.results[k], part.seen[k]
	}
	if b.steps != nil {
		next.steps, next.hashes = slices.Clone(b.steps), slices.Clone(b.hashes)
		for k, i := range at {
			next.steps[i], next.hashes[i] = part.steps[k], part.hashes[k]
		}
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
