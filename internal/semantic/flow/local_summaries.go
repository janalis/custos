package flow

import (
	"maps"

	"custos/internal/php/syntax"
	"custos/internal/semantic/infer"
)

type localSummaries struct {
	guard     func(func())
	base      *Snapshot
	inventory *summarizing
	order     []syntax.Node
	cache     map[string]Summary
	failed    map[syntax.Node]bool
	// alone remembers base.alone for the file, per contract key number of
	// the environment.
	alone aloneKeys
}

// NewLocal creates a file-query environment with demand-driven local contracts.
// Its overlay removes stale declarations immediately. Project publication still
// uses Summarize, which always visits the full declaration inventory. A non-nil
// guard runs deferred builds inside the caller's shared-data failure boundary.
func NewLocal(file *syntax.File, types *infer.Env, snapshot *Snapshot, options Options, guard func(func())) *Env {
	e, order := summaryInventory(file, types, snapshot, options)
	inventory := e.summarizing
	if inventory.base != nil {
		inventory.base = inventory.base.flat()
	}
	cache := maps.Clone(inventory.out.Summaries)
	placeholders := maps.Clone(cache)
	for key := range inventory.declared {
		placeholders[key] = Summary{}
	}
	mask := &File{Path: file.Path, options: e.options, Summaries: placeholders}
	e.snapshot = inventory.base.WithFile(mask)
	e.summarizing = nil
	// Every local lookup completes its stable dependency closure before answering.
	// This query view never changes; rebasing replaces it wholesale. Only workers
	// need read tracking for discovery and tentative recursive scope results.
	e.consulted = nil
	e.local = &localSummaries{guard: guard, base: inventory.base, inventory: inventory, order: order, cache: cache}
	return e
}

func (e *Env) localContract(key string) (Summary, bool, bool) {
	local := e.local
	if !local.alone.holds(e.keyTable().id(key), local.base, e.file.Path, key) {
		return Summary{}, false, false
	}
	if cached, known := local.cache[key]; known {
		return cached, true, true
	}
	declared := local.inventory.declared[key]
	if len(declared) != 1 {
		return Summary{}, false, false
	}
	if local.guard == nil {
		return e.materializeLocal(key, MaxDepth), true, true
	}
	var summary Summary
	local.guard(func() { summary = e.materializeLocal(key, MaxDepth) })
	return summary, true, true
}

// A dependency discovery pass finds the needed closure. Replaying only that
// closure in declaration order makes recursive contracts independent of which
// inspection first requested a member of the cycle. Cached closures are stable
// under the same immutable project snapshot and need not be recomputed.
func (e *Env) materializeLocal(key string, passes int) Summary {
	local := e.local
	if e.recordSites == nil {
		e.recordSites = map[recordSite]uint32{}
	}
	needed := map[syntax.Node]bool{local.inventory.declared[key][0]: true}
	for pass := 0; pass < passes; pass++ {
		worker := New(e.file, e.types, e.snapshot, e.options)
		worker.consulted = map[string]bool{}
		worker.keys = e.keyTable()
		worker.recordSites = e.recordSites
		worker.sharing = e.sharing
		worker.localBodies = e.localBodies
		worker.declarations, worker.classKeys, worker.kindOf = e.declarations, e.classKeys, e.kindOf
		if e.globalsReady {
			worker.globalsKnown, worker.globalsReady = true, true
			worker.global, worker.globalTaint, worker.globalReads = e.global, e.globalTaint, e.globalReads
		}
		out := &File{Path: e.file.Path, options: e.options, Summaries: maps.Clone(local.inventory.out.Summaries)}
		state := &summarizing{base: local.base, cached: local.cache, borrowed: e.scopes, out: out, declared: local.inventory.declared, names: local.inventory.names, state: map[syntax.Node]uint8{}, dependencies: map[syntax.Node]*readSet{}}
		worker.summarizing = state
		var first syntax.Node
		for _, node := range local.order {
			if needed[node] {
				if _, cached := local.cache[state.names[node]]; cached {
					continue
				}
				if first == nil {
					first = node
				}
				worker.summarize(node)
			}
		}
		e.declarations, e.classKeys, e.kindOf = worker.declarations, worker.classKeys, worker.kindOf
		e.sharing = worker.sharing
		stable := true
		for node := range state.state {
			if !needed[node] {
				needed[node] = true
				stable = false
			}
		}
		if !stable {
			recursive := false
			for id, read := range worker.answers {
				name := worker.keys.key(uint32(id))
				recursive = recursive || read.held && len(state.declared[name]) == 1 && state.base.alone(e.file.Path, name)
			}
			if recursive {
				for _, node := range local.order {
					if _, cached := local.cache[state.names[node]]; needed[node] && !cached {
						recursive = node != first
						break
					}
				}
			}
			if recursive {
				unstable := unstableLocal(state, worker.keys, worker.answers, e.file.Path)
				for node := range state.state {
					name := state.names[node]
					if id, known := worker.keys.known(name); !known || !unstable.has(id) {
						local.cache[name] = out.Summaries[name]
					}
				}
				for scope, result := range worker.scopes {
					if _, exists := e.scopes[scope]; !exists && result.reads != nil && !result.reads.meets(unstable) {
						e.scopes[scope] = result
					}
				}
				continue
			}
		}
		// Contracts stay private to this environment; the immutable mask never
		// changes. Reuse only scopes whose recorded answers still match the cache.
		for name, summary := range out.Summaries {
			local.cache[name] = summary
		}
		// Own answers are immutable once served, and the external snapshot is
		// unchanged throughout this worker. Only a held local recursion hole
		// can differ when the completed cache is published.
		differ := &readSet{}
		for id, read := range worker.answers {
			if name := worker.keys.key(uint32(id)); read.held && len(state.declared[name]) == 1 && local.base.alone(e.file.Path, name) {
				differ.add(uint32(id))
			}
		}
		for scope, result := range worker.scopes {
			if _, exists := e.scopes[scope]; !exists && result.reads != nil && !result.reads.meets(differ) {
				e.scopes[scope] = result
			}
		}
		return local.cache[key]
	}
	if local.failed == nil {
		local.failed = map[syntax.Node]bool{}
	}
	local.failed[local.inventory.declared[key][0]] = true
	local.cache[key] = Summary{}
	return Summary{}
}

// A recursive hole can change only contracts that transitively read it. Other
// completed components and their body facts are stable across canonical replay.
func unstableLocal(state *summarizing, keys *contractKeys, answers []answer, path string) *readSet {
	bad := &readSet{}
	for id, read := range answers {
		if name := keys.key(uint32(id)); read.held && len(state.declared[name]) == 1 && state.base.alone(path, name) {
			bad.add(uint32(id))
		}
	}
	for changed := true; changed; {
		changed = false
		for node, reads := range state.dependencies {
			id := keys.id(state.names[node])
			if !bad.has(id) && reads.meets(bad) {
				bad.add(id)
				changed = true
			}
		}
	}
	return bad
}
