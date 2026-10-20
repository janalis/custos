package lsp

import (
	"maps"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"custos/internal/php/syntax"
	"custos/internal/project"
	"custos/internal/semantic/flow"
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
func (s *Server) buildIndex() {
	s.mu.Lock()
	e, cfg, generation := s.engine, s.cfg, s.indexGeneration
	s.mu.Unlock()
	start := time.Now()
	var paths []string
	for _, p := range cfg.Paths {
		paths = append(paths, filepath.Join(cfg.Root, p))
	}
	files, err := discoverIndexFiles(paths, cfg.Exclude)
	if err != nil {
		s.logf("index: %v", err)
		s.mu.Lock()
		if s.indexGeneration == generation && s.engine == e {
			s.indexing, s.pending = false, nil
		}
		s.mu.Unlock()
		return
	}
	ix := buildIndexSymbols(project.IndexSources(cfg.Root, files), syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}, infer.Options{Assertions: cfg.Flow.Assertions})
	var snapshot *flow.Snapshot
	if e.NeedsFlow() {
		snapshot = project.BuildFlowSnapshot(files, nil, syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag}, ix, cfg.Flow)
	}
	s.mu.Lock()
	if s.engine != e || s.indexGeneration != generation {
		s.mu.Unlock()
		return
	}
	s.engine = e.WithIndex(ix)
	if snapshot != nil {
		s.engine = s.engine.WithFlow(snapshot)
		s.flow = snapshot
	}
	s.index = ix
	s.indexing = false
	queued := s.pending
	s.pending = nil
	s.mu.Unlock()
	if len(queued) > 0 {
		s.watchedFilesChanged(queued) // replays and re-analyses
	}
	if e.NeedsFlow() {
		s.refreshFlow()
	}
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
	build := project.BuildFlow(files, sources, opt, ix, cfg.Flow)
	s.mu.Lock()
	if s.engine == e {
		s.flow, s.flowBuild, s.flowSymbols = build.Snapshot, build, built
		s.engine = e.WithFlow(build.Snapshot)
	}
	s.mu.Unlock()
}

// Test seams also keep each build's filesystem/index operations explicit.
var (
	discoverIndexFiles = project.Discover
	buildIndexSymbols  = project.BuildIndex
	updateFlow         = (*project.FlowBuild).Update
)
