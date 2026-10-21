package lsp

import (
	"encoding/json"
	"fmt"
	"maps"
	"path/filepath"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/catalogue"
	"custos/internal/project/baseline"
	"custos/internal/project/config"
)

func (s *Server) initialize(raw json.RawMessage) (any, *rpcError) {
	var p initializeParams
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, &rpcError{Code: codeInvalidParams, Message: err.Error()}
	}
	s.root = "."
	switch {
	case len(p.WorkspaceFolders) > 0:
		s.root = uriToPath(p.WorkspaceFolders[0].URI)
	case p.RootURI != "":
		s.root = uriToPath(p.RootURI)
	case p.RootPath != "":
		s.root = p.RootPath
	}
	s.initOpts = p.InitializationOptions
	s.resolve = p.Capabilities.TextDocument.CodeAction.ResolveSupport != nil
	s.watchDynamic = p.Capabilities.Workspace.DidChangeWatchedFiles.DynamicRegistration
	if err := s.configure(); err != nil {
		return nil, &rpcError{Code: codeRequestFailed, Message: err.Error()}
	}
	return map[string]any{
		"capabilities": map[string]any{
			"positionEncoding": "utf-16",
			"textDocumentSync": map[string]any{"openClose": true, "change": 2, "save": map[string]any{"includeText": false}},
			"codeActionProvider": map[string]any{
				"codeActionKinds": []string{"quickfix", kindFixAll},
				"resolveProvider": true,
			},
			"executeCommandProvider": map[string]any{"commands": []string{cmdFixFile, cmdFixRule}},
		},
		"serverInfo": map[string]any{"name": "custos", "version": Version},
	}, nil
}

// registry returns the rules served (overridable in tests).
var registry = catalogue.All

// Version is reported in serverInfo (set by main).
var Version = "dev"

// configure (re)builds the engine from custos.json + initializationOptions.
func (s *Server) configure() error {
	cfg, err := config.Load(s.root)
	if err != nil {
		return err
	}
	if len(s.initOpts) > 0 && string(s.initOpts) != "null" {
		var over config.File
		if err := json.Unmarshal(s.initOpts, &over); err != nil {
			return fmt.Errorf("initializationOptions: %w", err)
		}
		if cfg, err = mergeOverrides(cfg, over); err != nil {
			return err
		}
	}
	e, err := analysis.NewEngine(registry(), cfg.Analysis())
	if err != nil {
		return err
	}
	bl := s.loadBaseline(cfg)
	s.mu.Lock()
	modeChanged := s.cfg != nil && s.cfg.Flow != cfg.Flow
	restart := modeChanged || s.indexing
	if restart {
		s.indexGeneration++
		s.indexing = false
		if !e.NeedsIndex() {
			s.pending = nil
		}
	}
	if modeChanged {
		s.index = nil
		s.flow = nil
	}
	// Flow options may have changed: the next edit rebuilds the summaries.
	s.flowBuild, s.flowSymbols = nil, nil
	if s.index != nil {
		e = e.WithIndex(s.index)
		if s.flow != nil {
			e = e.WithFlow(s.flow)
		}
	}
	s.cfg, s.engine, s.baseline = cfg, e, bl
	s.mu.Unlock()
	if restart && s.beginIndexing() {
		go s.buildIndex()
	}
	return nil
}

// baselinePath returns the absolute path of cfg's baseline, "" without one
// (or without a configuration, when the project's failed to load).
func baselinePath(cfg *config.Config) string {
	if cfg == nil || cfg.Baseline == "" {
		return ""
	}
	return filepath.Join(cfg.Root, cfg.Baseline)
}

// loadBaseline reads cfg's baseline. An unreadable one is reported and
// ignored, so the editor still shows every finding rather than none.
func (s *Server) loadBaseline(cfg *config.Config) *baseline.Baseline {
	path := baselinePath(cfg)
	if path == "" {
		return nil
	}
	bl, err := loadBaselineFile(path)
	if err != nil {
		s.logf("baseline: %v", err)
		return nil
	}
	return bl
}

// loadBaselineFile reads a baseline file (a test seam).
var loadBaselineFile = baseline.Load

// reloadBaseline rereads the baseline after its file changed on disk.
func (s *Server) reloadBaseline() {
	s.mu.Lock()
	cfg := s.cfg
	s.mu.Unlock()
	// Read without the lock (a baseline may be large); a reconfiguration
	// meanwhile already installed its own baseline.
	bl := s.loadBaseline(cfg)
	s.mu.Lock()
	stale := s.cfg != cfg
	if !stale {
		s.baseline = bl
	}
	s.mu.Unlock()
	if !stale {
		s.reanalyzeAll()
	}
}

// mergeOverrides applies client-provided settings on top of the project config.
func mergeOverrides(base *config.Config, over config.File) (*config.Config, error) {
	f := config.File{Flow: &config.FlowSettings{}, PHP: over.PHP, ComparisonStyle: over.ComparisonStyle, ShortOpenTag: over.ShortOpenTag, Rules: over.Rules}
	if f.PHP == "" {
		f.PHP = base.PHP.String()
	}
	if f.ComparisonStyle == "" && base.ComparisonStyle == analysis.StyleYoda {
		f.ComparisonStyle = "yoda"
	}
	if over.Flow != nil {
		*f.Flow = *over.Flow
	}
	if f.Flow.Assertions == nil {
		f.Flow.Assertions = &base.Flow.Assertions
	}
	if f.ShortOpenTag == nil {
		f.ShortOpenTag = &base.ShortOpenTag
	}
	merged, err := config.Resolve(base.Root, f)
	if err != nil {
		return nil, err
	}
	// What to analyse stays the project's choice.
	merged.Paths, merged.Exclude, merged.Baseline = base.Paths, base.Exclude, base.Baseline
	// A client rule entry overrides the project's field by field.
	for id, rc := range base.Rules {
		over, ok := merged.Rules[id]
		if !ok {
			merged.Rules[id] = rc
			continue
		}
		if over.Enabled == nil {
			over.Enabled = rc.Enabled
		}
		if over.Severity == "" {
			over.Severity = rc.Severity
		}
		opts := maps.Clone(rc.Options)
		if opts == nil {
			opts = map[string]any{}
		}
		maps.Copy(opts, over.Options)
		over.Options = opts
		merged.Rules[id] = over
	}
	return merged, nil
}
