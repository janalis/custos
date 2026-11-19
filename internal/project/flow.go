package project

import (
	"context"
	"hash/maphash"
	"maps"
	"runtime"
	"runtime/debug"
	"slices"
	"sync/atomic"

	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/flow"
	"github.com/janalis/custos/internal/semantic/index"
	"github.com/janalis/custos/internal/semantic/infer"
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
	// sizes holds the length of each file's source as the build read it,
	// nil when unknown (see summarize).
	sizes []int
	// steps holds, when the build records them, every summary of each file
	// in the order the build made them, with the snapshot it was made
	// against; hashes holds the hash of each file's source. A later build
	// of the same source takes them as priors (see reuse).
	steps  [][]flowStep
	hashes []uint64
	// priors holds, per file, the steps an earlier build made from the same
	// source (see seed).
	priors [][]flowStep
	// made counts the summaries the build made instead of taking them from
	// its priors.
	made atomic.Int64
	// kept holds, while a build given the trees of its files (--parse-once)
	// that records no steps composes, the inference environment (with the
	// tree and names it read) of the last summary of each file, which its next summary infers with instead of
	// parsing and inferring again; keptBytes estimates what they hold,
	// bounded by keptLimit (see keepLimit).
	kept      []*infer.Env
	keptBytes atomic.Int64
	keptLimit int64
}

// flowStep is one summary a build made of a file, with the snapshot it
// was made against.
type flowStep struct {
	result  *flow.File
	against *flow.Snapshot
}

// sourceSeed keys the source hashes of every build of this process.
var sourceSeed = maphash.MakeSeed()

// Files returns the files the build summarized, in build order.
func (b *FlowBuild) Files() []string { return b.files }

// FromDisk reports whether the build summarized path as read from disk: it
// is one of its files and no source was supplied for it.
func (b *FlowBuild) FromDisk(path string) bool {
	_, file := b.at[path]
	_, supplied := b.sources[path]
	return file && !supplied
}

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
// of being summarized again once they settle. Cold builds discover the
// dependency order during composition; the supplied order is only a hint
// that can avoid early repeated summaries.
func BuildFlowOrdered(ctx context.Context, files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options flow.Options, order map[string]int) *FlowBuild {
	return buildFlow(ctx, files, srcs, opt, ix, options, order, nil, false, nil)
}

// BuildFlowRecorded is BuildFlowOrdered keeping every summary it makes, so
// that Update, when an edit within a recursion makes it summarize files
// from none again, takes those the files' contracts still support instead
// of summarizing them (see reuse). The language server, which updates the
// build on every pause in editing, takes it; it costs memory a single run
// does not spend.
func BuildFlowRecorded(ctx context.Context, files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options flow.Options, order map[string]int) *FlowBuild {
	return buildFlow(ctx, files, srcs, opt, ix, options, order, nil, true, nil)
}

// buildFlow is BuildFlowOrdered taking, for each file whose source is the
// one earlier read, the summaries earlier made of it as priors, and
// recording its own steps when record is set. A file with a tree in trees
// (parsed from its source in srcs) is summarized from it instead of being
// parsed again.
func buildFlow(ctx context.Context, files []string, srcs [][]byte, opt syntax.Options, ix *index.Index, options flow.Options, order map[string]int, earlier *FlowBuild, record bool, trees []*syntax.File) *FlowBuild {
	// Sources are read once. Without trees, each summary parses its file
	// again and keeps no tree: compose then holds the trees of the files
	// being summarized, not those of every file, which take about ten times
	// their source. With trees, the inference environment of a file's last
	// summary is kept for its next one, within a budget (see keep). A file
	// that cannot be read or crashes the parser is not summarized.
	read := make([][]byte, len(files))
	readable := make([]bool, len(files))
	_ = parallel(ctx, len(files), 0, func(i int) {
		var src []byte
		if i < len(srcs) {
			src = srcs[i]
		}
		if src == nil {
			var err error
			if src, err = ReadSource(files[i]); err != nil {
				return
			}
		}
		read[i], readable[i] = src, true
	}, nil)
	b := &FlowBuild{files: files, at: make(map[string]int, len(files)), results: make([]*flow.File, len(files)), sources: map[string][]byte{}, opt: opt, options: options, order: order, traced: true, seen: make([]map[string]bool, len(files)), declares: make([][]string, len(files))}
	for i, path := range files {
		b.at[path] = i
		if i < len(srcs) && srcs[i] != nil {
			b.sources[path] = srcs[i]
		}
	}
	b.sizes = make([]int, len(files))
	for i, src := range read {
		b.sizes[i] = len(src)
	}
	if record {
		b.steps, b.hashes = make([][]flowStep, len(files)), make([]uint64, len(files))
		for i, src := range read {
			b.hashes[i] = maphash.Bytes(sourceSeed, src)
		}
		b.seed(earlier, readable)
	}
	tree := func(i int) *syntax.File {
		if i < len(trees) && trees[i] != nil {
			return trees[i]
		}
		if !readable[i] {
			return nil
		}
		return syntax.ParseBest(files[i], read[i], opt)
	}
	snapshot := flow.NewSnapshotWithOptions(options)
	against := make([]*flow.Snapshot, len(files))
	// With the trees of --parse-once, the inference environments built on
	// them are kept too, from the first summary on: that run already holds
	// every analyzed file's tree, and trades memory for time.
	if !record && trees != nil {
		b.kept, b.keptLimit = make([]*infer.Env, len(files)), keepLimit()
	}
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
	// No summary follows: the environments kept for one are released.
	b.kept = nil
	// The earlier build's steps are no longer needed: the build now
	// published must not keep them alive.
	b.priors = nil
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
	cyclic := make([]bool, len(edges))
	for _, component := range flowComponents(edges, in) {
		for _, i := range component {
			cyclic[i] = len(component) > 1 || slices.Contains(edges[i], i)
		}
	}
	return cyclic
}

// flowComponents returns strongly connected components, dependencies first.
func flowComponents(edges [][]int, in []bool) [][]int {
	n := len(edges)
	index, low := make([]int, n), make([]int, n)
	for i := range index {
		index[i] = -1
	}
	onStack := make([]bool, n)
	var components [][]int
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
			components = append(components, component)
		}
	}
	for v := range n {
		if in[v] && index[v] < 0 {
			connect(v)
		}
	}
	return components
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
// whose summaries read a changed contract. Each round discovers a new
// dependency graph; chain length within that graph does not consume rounds.
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
// wave read one immutable snapshot and run side by side; dependency waves
// run in order. What has not settled when it stops is dropped (see unsettled).
func (b *FlowBuild) compose(ctx context.Context, snapshot *flow.Snapshot, against []*flow.Snapshot, ix *index.Index, tree func(int) *syntax.File) *flow.Snapshot {
	summarized := make([]int, len(b.results))
	for range maxRounds {
		changed := false
		for _, wave := range b.flowWaves() {
			// Acyclic files settle in one pass after their dependencies.
			// Recursion settles within its wave before its readers run.
			for range maxSummaries {
				if ctx.Err() != nil {
					return snapshot
				}
				var work []int
				stale := b.outdated(wave.files, snapshot, against)
				for k, i := range wave.files {
					if stale[k] && summarized[i] < maxSummaries {
						work = append(work, i)
						summarized[i]++
					}
				}
				if len(work) == 0 {
					break
				}
				names := b.declared(work)
				prev := b.resultsOf(work)
				b.summarize(ctx, work, snapshot, against, ix, tree)
				for _, i := range work {
					if summarized[i] == maxSummaries {
						b.take(i)
					}
				}
				next := snapshot.Replace(prev, b.resultsOf(work))
				// Only the files of work changed: the other contracts are
				// those of snapshot.
				if snapshot.EqualFor(next, b.declared(work, names...)) {
					snapshot = next
					break
				}
				snapshot, changed = next, true
				if !wave.recursive {
					break
				}
			}
		}
		if !changed {
			break
		}
	}

	return b.unsettled(snapshot, against)
}

// flowWaves groups dependency components by depth. Independent components
// remain parallel; cycles share a wave and use compose's per-file budget.
type flowWave struct {
	files     []int
	recursive bool
}

func (b *FlowBuild) flowWaves() []flowWave {
	edges := b.reads()
	in := make([]bool, len(edges))
	for i := range in {
		in[i] = true
	}
	components := flowComponents(edges, in)
	owner := make([]int, len(edges))
	for c, files := range components {
		for _, i := range files {
			owner[i] = c
		}
	}
	depth := make([]int, len(components))
	var waves []flowWave
	for c, files := range components {
		for _, i := range files {
			for _, j := range edges[i] {
				if owner[j] != c {
					depth[c] = max(depth[c], depth[owner[j]]+1)
				}
			}
		}
		for len(waves) <= depth[c] {
			waves = append(waves, flowWave{})
		}
		wave := &waves[depth[c]]
		wave.files = append(wave.files, files...)
		wave.recursive = wave.recursive || len(files) > 1 || slices.Contains(edges[files[0]], files[0])
	}
	return waves
}

// outdated reports, for each file of files, that it has no summaries or that
// a contract they read differs from the snapshot they were computed against.
// The checks only read immutable snapshots, so they run side by side.
func (b *FlowBuild) outdated(files []int, snapshot *flow.Snapshot, against []*flow.Snapshot) []bool {
	out := make([]bool, len(files))
	check := func(k int) {
		r := b.results[files[k]]
		out[k] = r == nil || !r.Current(against[files[k]], snapshot)
	}
	if len(files) < minParallelChecks {
		for k := range files {
			check(k)
		}
		return out
	}
	_ = parallel(context.Background(), len(files), 0, check, nil)
	return out
}

// minParallelChecks is the number of files from which outdated checks them on
// every CPU: a few checks cost less than starting the workers.
const minParallelChecks = 64

// declared appends to names the names the current summaries of the files
// of work declare.
func (b *FlowBuild) declared(work []int, names ...string) []string {
	for _, i := range work {
		if r := b.results[i]; r != nil {
			for name := range r.Summaries {
				names = append(names, name)
			}
		}
	}
	return names
}

// summarize summarizes the files of work side by side against snapshot,
// the largest sources first: a wave lasts at least as long as its slowest
// file, which should not start last.
func (b *FlowBuild) summarize(ctx context.Context, work []int, snapshot *flow.Snapshot, against []*flow.Snapshot, ix *index.Index, tree func(int) *syntax.File) {
	if b.sizes != nil {
		work = slices.Clone(work)
		slices.SortStableFunc(work, func(x, y int) int { return b.sizes[y] - b.sizes[x] })
	}
	_ = parallel(ctx, len(work), summaryWorkers(runtime.GOMAXPROCS(0)), func(j int) {
		i := work[j]
		var basis *flow.Snapshot
		if step, ok := b.reuse(i, snapshot); ok {
			b.results[i], basis = step.result, step.against
		} else if types := b.types(i, ix, tree); types != nil {
			b.made.Add(1)
			// The file's first summary starts from the first one the
			// earlier build made, if any, which read the same contracts
			// the closest.
			prior, earlier := b.results[i], against[i]
			if prior == nil && len(b.priors) > 0 && len(b.priors[i]) > 0 {
				prior, earlier = b.priors[i][0].result, b.priors[i][0].against
			}
			b.results[i] = extractFlow(types.File, ix, b.opt.Version, snapshot, b.options, prior, earlier, types)
			b.keep(i, types, snapshot)
		} else {
			return
		}
		against[i] = snapshot
		if b.steps != nil && b.results[i] != nil {
			// A reused summary keeps the basis it was recorded with: it
			// answers what it consulted as snapshot does.
			if basis == nil {
				basis = b.results[i].Basis(snapshot)
			}
			b.steps[i] = append(b.steps[i], flowStep{b.results[i], basis})
		}
		if b.traced && b.results[i] != nil {
			b.trace(i)
		}
	}, nil)
}

// types returns the inference environment the summary of file i infers
// with: the one its last summary kept, else a new one over its tree; nil
// when it has no tree.
func (b *FlowBuild) types(i int, ix *index.Index, tree func(int) *syntax.File) *infer.Env {
	if types := b.take(i); types != nil {
		types.Renew()
		return types
	}
	file := tree(i)
	if file == nil {
		return nil
	}
	return flow.NewTypes(file, ix, b.opt.Version, b.options)
}

// keep keeps types, which just summarized file i, for its next summary.
// Only summaries that read another file's contracts can be outdated, so
// only their environments are kept, while the kept ones fit keptLimit:
// each holds the file's tree, names and inferred types, several times its
// source.
func (b *FlowBuild) keep(i int, types *infer.Env, snapshot *flow.Snapshot) {
	if b.kept == nil || b.results[i] == nil || !b.results[i].ReadsOthers(snapshot) {
		return
	}
	// Reserved with a compare-and-swap: concurrent workers near the limit
	// must not refuse each other's environments.
	cost := keptCost(b.sizes[i])
	for {
		held := b.keptBytes.Load()
		if held+cost > b.keptLimit {
			return
		}
		if b.keptBytes.CompareAndSwap(held, held+cost) {
			break
		}
	}
	b.kept[i] = types
}

// take returns and forgets the environment kept for file i, nil for none.
func (b *FlowBuild) take(i int) *infer.Env {
	if b.kept == nil || b.kept[i] == nil {
		return nil
	}
	types := b.kept[i]
	b.kept[i] = nil
	b.keptBytes.Add(-keptCost(b.sizes[i]))
	return types
}

// keptCost estimates the live memory the environment of a file of size
// bytes holds after a summary: on laravel, its tree, tokens and inferred
// types took about 28 times its source.
func keptCost(size int) int64 { return int64(size) * 28 }

// keepBudget bounds the estimated memory of the environments kept at once:
// laravel's take about 140 MiB. A soft memory limit (see debug.SetMemoryLimit)
// lowers it to an eighth of that limit.
var keepBudget int64 = 256 << 20

// keepLimit returns keepBudget within the soft memory limit.
func keepLimit() int64 {
	return min(keepBudget, debug.SetMemoryLimit(-1)/8)
}

// seed takes as priors the steps earlier recorded for the files whose
// source this build reads unchanged (the same hash).
func (b *FlowBuild) seed(earlier *FlowBuild, readable []bool) {
	if earlier == nil {
		return
	}
	b.priors = make([][]flowStep, len(b.files))
	for i, path := range b.files {
		if j, ok := earlier.at[path]; ok && readable[i] && earlier.hashes[j] == b.hashes[i] {
			b.priors[i] = earlier.steps[j]
		}
	}
}

// reuse returns a step an earlier build made of file i from the same
// source whose summary is current against snapshot; ok is false when there
// is none.
func (b *FlowBuild) reuse(i int, snapshot *flow.Snapshot) (step flowStep, ok bool) {
	if len(b.priors) == 0 {
		return flowStep{}, false
	}
	for _, step := range b.priors[i] {
		if step.result.Current(step.against, snapshot) {
			return step, true
		}
	}
	return flowStep{}, false
}

// resultsOf returns the current summaries of the files of work.
func (b *FlowBuild) resultsOf(work []int) []*flow.File {
	out := make([]*flow.File, len(work))
	for k, i := range work {
		out[k] = b.results[i]
	}
	return out
}

// summaryWorkers is the number of files summarized side by side on cores
// CPUs. Summaries allocate heavily: on symfony/src (16 cores) every core
// made the whole run about 8% slower, so many cores use half. A small
// machine keeps them all: on 4 cores, 2 workers made laravel's run about
// 6% slower.
func summaryWorkers(cores int) int {
	return max(cores/2, min(cores, 8))
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
		var held []int
		for i, r := range b.results {
			if r != nil {
				held = append(held, i)
			}
		}
		for k, stale := range b.outdated(held, snapshot, against) {
			if stale {
				b.results[held[k]] = nil
				dropped = true
			}
		}
		if !dropped {
			return snapshot
		}
		snapshot = flow.NewSnapshotWithOptions(b.options, b.results...)
	}
}

// extractFlow summarizes a file, again when prior holds its summaries
// computed against earlier (see flow.Reextract), inferring types with
// types, a new environment when nil (see flow.NewTypes); tests replace it to
// simulate a crash. A file whose summaries crash keeps none: calls into it
// read as unresolved.
var extractFlow = func(file *syntax.File, ix *index.Index, php phpversion.Version, snapshot *flow.Snapshot, options flow.Options, prior *flow.File, earlier *flow.Snapshot, types *infer.Env) *flow.File {
	if types == nil {
		types = flow.NewTypes(file, ix, php, options)
	}
	if prior != nil && earlier != nil {
		return flow.ReextractWith(file, types, snapshot, options, prior, earlier)
	}
	return flow.ExtractWith(file, types, snapshot, options)
}
