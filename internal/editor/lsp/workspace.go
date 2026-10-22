package lsp

import (
	"context"
	"maps"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/project"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
)

// beginIndexing reports whether an enabled rule needs cross-file symbols
// and, if so, marks the index as being built: file changes arriving from
// now on are queued for buildIndex instead of being dropped.
func (s *Server) beginIndexing() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.engine == nil || !s.engine.NeedsIndex() {
		return false
	}
	s.indexing = true
	s.indexGeneration++
	return true
}

// buildIndex indexes the workspace (and vendor) after beginIndexing, then
// re-analyses open documents.
// abandonIndex ends a failed index build, unless a later one replaced it:
// analyses go on without a project index.
func (s *Server) abandonIndex(e *analysis.Engine, generation uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.indexGeneration == generation && s.engine == e {
		s.indexing, s.pending = false, nil
	}
}

func (s *Server) buildIndex() {
	s.mu.Lock()
	e, cfg, generation := s.engine, s.cfg, s.indexGeneration
	s.mu.Unlock()
	// It runs on its own goroutine, outside the request guards.
	defer s.guard("index", func() { s.abandonIndex(e, generation) })
	start := time.Now()
	var paths []string
	for _, p := range cfg.Paths {
		paths = append(paths, filepath.Join(cfg.Root, p))
	}
	files, err := discoverIndexFiles(paths, cfg.Exclude)
	if err != nil {
		s.logf("index: %v", err)
		s.abandonIndex(e, generation)
		return
	}
	ix := buildIndexSymbols(cfg.Root, project.IndexSources(cfg.Root, files, paths, cfg.Exclude), syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}, infer.Options{Assertions: cfg.Flow.Assertions})
	s.mu.Lock()
	if s.engine != e || s.indexGeneration != generation {
		s.mu.Unlock()
		return
	}
	s.engine = e.WithIndex(ix)
	s.index = ix
	s.indexing = false
	queued := s.pending
	s.pending = nil
	s.mu.Unlock()
	if len(queued) > 0 {
		s.watchedFilesChanged(queued) // replays and re-analyses
	}
	// One flow build, over the final index and the open buffers.
	s.refreshFlow()
	n, _, _, _ := ix.Stats()
	s.logf("indexed %d files in %v", n, time.Since(start).Round(time.Millisecond))
	s.reanalyzeAll()
}

// fileChange is one workspace/didChangeWatchedFiles event.
type fileChange struct {
	URI  string `json:"uri"`
	Type int    `json:"type"` // 1 created, 2 changed, 3 deleted
}

// watchedFilesChanged updates the project index for files changed on disk
// and re-analyses open documents (their semantic findings may depend on them).
// Changes arriving while the initial index is built are queued and replayed.
func (s *Server) watchedFilesChanged(changes []fileChange) {
	s.mu.Lock()
	bl := baselinePath(s.cfg)
	s.mu.Unlock()
	for _, ch := range changes {
		if bl != "" && filepath.Clean(uriToPath(ch.URI)) == bl {
			s.reloadBaseline()
			break
		}
	}
	s.mu.Lock()
	if s.indexing {
		s.pending = append(s.pending, changes...)
		s.mu.Unlock()
		return
	}
	ix, cfg := s.index, s.cfg
	s.mu.Unlock()
	if ix == nil || len(changes) == 0 {
		return
	}
	touched := false
	for _, ch := range changes {
		path := uriToPath(ch.URI)
		if !strings.HasSuffix(strings.ToLower(path), ".php") {
			continue
		}
		touched = true
		if ch.Type == 3 {
			ix.Remove(path)
			continue
		}
		src, err := project.ReadSource(path)
		if err != nil {
			ix.Remove(path)
			continue
		}
		fs := project.ExtractSymbols(path, src, syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}, infer.Options{Assertions: cfg.Flow.Assertions})
		ix.DropStaleInferred(fs)
		ix.Add(fs)
	}
	if touched {
		s.refreshFlow()
		s.reanalyzeAll()
	}
}

// reindexDoc refreshes the index entry of an open document from its buffer
// and the flow summaries; it reports whether other documents may need
// analysing again (false when only the document's own findings can have
// changed).
func (s *Server) reindexDoc(uri string) bool {
	s.mu.Lock()
	d, ok := s.docs[uri]
	ix, cfg, build := s.index, s.cfg, s.flowBuild
	var text []byte
	var path string
	var prev *index.FileSymbols
	if ok {
		text, path, prev = d.text, d.path, s.flowSymbols[d.path]
	}
	s.mu.Unlock()
	if !ok || ix == nil {
		return false
	}
	fs := project.ExtractSymbols(path, text, syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}, infer.Options{Assertions: cfg.Flow.Assertions})
	ix.DropStaleInferred(fs)
	ix.Add(fs)
	if build != nil && prev != nil && index.SameDeclarations(prev, fs) {
		// No other file sees a change: summarize again only what the
		// buffer's summaries reach.
		if next, ok := updateFlow(build, map[string][]byte{path: text}, ix); ok {
			s.mu.Lock()
			if s.flowBuild == build {
				changed := !next.Snapshot.Equal(build.Snapshot)
				symbols := maps.Clone(s.flowSymbols)
				symbols[path] = fs
				s.flowBuild, s.flowSymbols, s.flow = next, symbols, next.Snapshot
				if changed {
					s.engine = s.engine.WithFlow(next.Snapshot)
				}
				s.mu.Unlock()
				return changed
			}
			s.mu.Unlock()
		}
	}
	s.refreshFlow()
	return true
}

// refreshFlow rebuilds only project-local summaries and overlays open buffers.
// Publish the immutable snapshot through an engine copy under the server lock.
func (s *Server) refreshFlow() {
	s.mu.Lock()
	e, ix, cfg := s.engine, s.index, s.cfg
	if e == nil || !e.NeedsFlow() || ix == nil {
		s.mu.Unlock()
		return
	}
	type buffer struct {
		path string
		src  []byte
	}
	buffers := []buffer{}
	for _, d := range s.docs {
		buffers = append(buffers, buffer{d.path, d.text})
	}
	s.mu.Unlock()
	sort.Slice(buffers, func(i, j int) bool { return buffers[i].path < buffers[j].path })
	built := make(map[string]*index.FileSymbols, len(buffers))
	for _, b := range buffers {
		symbols := project.ExtractSymbols(b.path, b.src, syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}, infer.Options{Assertions: cfg.Flow.Assertions})
		ix.DropStaleInferred(symbols)
		ix.Add(symbols)
		built[b.path] = symbols
	}
	paths := []string{}
	for _, p := range cfg.Paths {
		paths = append(paths, filepath.Join(cfg.Root, p))
	}
	files, err := project.Discover(paths, cfg.Exclude)
	if err != nil {
		return
	}
	opt := syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}
	contents := make(map[string][]byte, len(buffers))
	present := make(map[string]bool, len(files))
	for _, path := range files {
		present[path] = true
	}
	for _, b := range buffers {
		contents[b.path] = b.src
		if !present[b.path] {
			files = append(files, b.path)
			present[b.path] = true
		}
	}
	sort.Strings(files)
	sources := make([][]byte, len(files))
	for i, path := range files {
		sources[i] = contents[path]
	}
	// Compose every unsaved dependency in the same bounded fixed point.
	build := buildFlow(context.Background(), files, sources, opt, ix, cfg.Flow)
	s.mu.Lock()
	if s.engine == e {
		s.flow, s.flowBuild, s.flowSymbols = build.Snapshot, build, built
		s.engine = e.WithFlow(build.Snapshot)
	}
	s.mu.Unlock()
}

// indexWorkspace indexes files, taking the declarations of unchanged ones
// from the cache the command line shares (see project.SymbolCache).
func indexWorkspace(root string, files []string, opt syntax.Options, options infer.Options) *index.Index {
	var cache *project.SymbolCache
	if dir := project.CacheDir(); dir != "" {
		cache = project.OpenSymbolCache(dir, root, opt, options)
	}
	ix, _ := project.BuildIndexCached(files, 0, opt, options, cache)
	return ix
}

// Test seams also keep each build's filesystem/index operations explicit.
var (
	discoverIndexFiles = project.Discover
	buildIndexSymbols  = indexWorkspace
	buildFlow          = project.BuildFlow
	updateFlow         = (*project.FlowBuild).Update
)
