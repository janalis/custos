package project

import (
	"custos/internal/diagnostic"
	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
)

// Options contains resolved configuration; adapters own flags and settings.
type Options struct {
	Root           string
	Paths, Exclude []string
	Analysis       analysis.Config
	Parse          syntax.Options
}

// Project binds discovered files to an explicit inspection set.
type Project struct {
	Engine  *analysis.Engine
	Files   []string
	Options Options
}

// Open configures inspections and discovers their requested file patterns.
func Open(rules []analysis.Rule, opt Options) (*Project, error) {
	engine, err := analysis.NewEngine(rules, opt.Analysis)
	if err != nil {
		return nil, err
	}
	files, err := DiscoverWith(opt.Paths, opt.Exclude, engine.FilePatterns())
	if err != nil {
		return nil, err
	}
	return &Project{Engine: engine, Files: files, Options: opt}, nil
}

// AnalyzeReport indexes project and vendor declarations, reuses read sources,
// and releases lazy edit closures after analysis. Results retain input order.
func (p *Project) AnalyzeReport() []diagnostic.FileResult {
	engine := p.Engine
	var sources [][]byte
	if engine.NeedsIndex() {
		ix, srcs := BuildIndexKeep(IndexSources(p.Options.Root, p.Files), len(p.Files), p.Options.Parse)
		engine, sources = engine.WithIndex(ix), srcs
		if engine.NeedsFlow() {
			engine = engine.WithFlow(BuildFlowSnapshot(p.Files, srcs, p.Options.Parse, ix))
		}
	}
	return RunReport(engine, p.Files, sources, p.Options.Parse)
}

// PrepareFixes computes changes without writing; the adapter owns writes.
func (p *Project) PrepareFixes() []PreparedFix {
	engine := p.Engine
	if engine.NeedsIndex() {
		ix := BuildIndex(IndexSources(p.Options.Root, p.Files), p.Options.Parse)
		engine = engine.WithIndex(ix)
		if engine.NeedsFlow() {
			engine = engine.WithFlow(BuildFlowSnapshot(p.Files, nil, p.Options.Parse, ix))
		}
	}
	return PrepareFixes(engine, p.Files, p.Options.Parse)
}

// AnalyzeBuffer analyzes supplied bytes without reading the filesystem.
func AnalyzeBuffer(engine fixing.Analyzer, path string, src []byte, opt syntax.Options) diagnostic.FileResult {
	file := syntax.ParseBest(path, src, opt)
	return diagnostic.FileResult{Path: path, Src: src, Findings: engine.Analyze(file), Errors: file.Errors}
}
