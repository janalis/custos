package lsp

import (
	"path/filepath"
	"strings"
	"time"

	"custos/internal/php/syntax"
	"custos/internal/project"
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
	return true
}

// buildIndex indexes the workspace (and vendor) after beginIndexing, then
// re-analyses open documents.
func (s *Server) buildIndex() {
	s.mu.Lock()
	e, cfg := s.engine, s.cfg
	s.mu.Unlock()
	start := time.Now()
	var paths []string
	for _, p := range cfg.Paths {
		paths = append(paths, filepath.Join(cfg.Root, p))
	}
	files, err := project.Discover(paths, cfg.Exclude)
	if err != nil {
		s.logf("index: %v", err)
		s.mu.Lock()
		s.indexing, s.pending = false, nil
		s.mu.Unlock()
		return
	}
	ix := project.BuildIndex(project.IndexSources(cfg.Root, files), syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag})
	s.mu.Lock()
	if s.engine == e {
		s.engine = e.WithIndex(ix)
		s.index = ix
	}
	s.indexing = false
	queued := s.pending
	s.pending = nil
	s.mu.Unlock()
	if len(queued) > 0 {
		s.watchedFilesChanged(queued) // replays and re-analyses
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
		fs := project.ExtractSymbols(path, src, syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag})
		ix.DropStaleInferred(fs)
		ix.Add(fs)
	}
	if touched {
		s.reanalyzeAll()
	}
}

// reindexDoc refreshes the saved document's symbols in the project index.
// reindexDoc refreshes the index entry of an open document from its buffer;
// it reports whether the index changed.
func (s *Server) reindexDoc(uri string) bool {
	s.mu.Lock()
	d, ok := s.docs[uri]
	ix, cfg := s.index, s.cfg
	var text []byte
	var path string
	if ok {
		text, path = d.text, d.path
	}
	s.mu.Unlock()
	if !ok || ix == nil {
		return false
	}
	fs := project.ExtractSymbols(path, text, syntax.Options{Version: cfg.PHP, ShortOpenTag: cfg.ShortOpenTag})
	ix.DropStaleInferred(fs)
	ix.Add(fs)
	return true
}
