package lsp

import (
	"encoding/json"
	"fmt"

	"custos/internal/inspection/analysis"
	"custos/internal/inspection/catalogue"
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
	s.mu.Lock()
	if s.index != nil {
		e = e.WithIndex(s.index)
	}
	s.cfg, s.engine = cfg, e
	s.mu.Unlock()
	return nil
}

// mergeOverrides applies client-provided settings on top of the project config.
func mergeOverrides(base *config.Config, over config.File) (*config.Config, error) {
	f := config.File{PHP: over.PHP, ComparisonStyle: over.ComparisonStyle, ShortOpenTag: over.ShortOpenTag, Rules: over.Rules}
	if f.PHP == "" {
		f.PHP = base.PHP.String()
	}
	if f.ComparisonStyle == "" && base.ComparisonStyle == analysis.StyleYoda {
		f.ComparisonStyle = "yoda"
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
	for id, rc := range base.Rules {
		if _, ok := merged.Rules[id]; !ok {
			merged.Rules[id] = rc
		}
	}
	return merged, nil
}
