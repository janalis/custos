package project

import (
	"context"
	"path/filepath"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/fixing"
	"custos/internal/inspection/analysis"
	"custos/internal/php/syntax"
	"custos/internal/semantic/infer"
)

// Options contains resolved configuration; adapters own flags and settings.
type Options struct {
	Root           string
	Paths, Exclude []string
	// Scope lists the project paths whose PHP files are indexed for symbols
	// besides the analysed ones, so cross-file facts do not depend on which
	// files a run names. Empty means Paths.
	Scope    []string
	Analysis analysis.Config
	Parse    syntax.Options
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
	engine, sources := p.prepare()
	return RunReport(engine, p.Files, sources, p.Options.Parse)
}

// PrepareFixes computes changes without writing; the adapter owns writes.
func (p *Project) PrepareFixes() []PreparedFix {
	engine, sources := p.prepare()
	return PrepareFixes(engine, p.Files, sources, p.Options.Parse)
}

// prepare attaches the project index and flow snapshot the engine needs. It
// returns the project sources the index read (nil where unread), so analysis
// and fixes see the same bytes without reading each file again. Files that
// rules request by name, such as composer.json, are analysed but never
// indexed or summarized as PHP.
func (p *Project) prepare() (*analysis.Engine, [][]byte) {
	engine := p.Engine
	if !engine.NeedsIndex() {
		return engine, nil
	}
	scope := p.Options.Scope
	if len(scope) == 0 {
		scope = p.Options.Paths
	}
	var php []string
	var at []int
	patterns := engine.FilePatterns()
	for i, f := range p.Files {
		if strings.HasSuffix(f, ".php") || !matchesAny(filepath.Base(f), patterns) {
			php, at = append(php, f), append(at, i)
		}
	}
	ix, read := BuildIndexKeep(IndexSources(p.Options.Root, php, scope, p.Options.Exclude), len(php), p.Options.Parse, infer.Options{Assertions: p.Options.Analysis.Flow.Assertions})
	ix.Freeze()
	engine = engine.WithIndex(ix)
	if engine.NeedsFlow() {
		engine = engine.WithFlow(BuildFlowSnapshot(context.Background(), php, read, p.Options.Parse, ix, p.Options.Analysis.Flow))
	}
	sources := make([][]byte, len(p.Files))
	for j, i := range at {
		sources[i] = read[j]
	}
	return engine, sources
}

// AnalyzeBuffer analyzes supplied bytes without reading the filesystem.
func AnalyzeBuffer(engine fixing.Analyzer, path string, src []byte, opt syntax.Options) (result diagnostic.FileResult) {
	defer func() {
		if r := recover(); r != nil {
			result = internalFailure(path, src, r)
		}
	}()
	file := syntax.ParseBest(path, src, opt)
	result = diagnostic.FileResult{Path: path, Src: src, Errors: file.Errors}
	if detailed, ok := engine.(interface {
		AnalyzeResult(*syntax.File) diagnostic.AnalysisResult
	}); ok {
		analysis := detailed.AnalyzeResult(file)
		result.Findings, result.Incomplete = analysis.Findings, analysis.Incomplete
	} else {
		result.Findings = engine.Analyze(file)
		if len(file.Errors) > 0 {
			result.Incomplete = []diagnostic.AnalysisIssue{{Reason: "syntax-recovery", Span: file.Errors[0].Span}}
		}
	}
	return result
}
