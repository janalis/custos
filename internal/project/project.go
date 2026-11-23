package project

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/fixing"
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/php/syntax"
	"github.com/janalis/custos/internal/semantic/flow"
	"github.com/janalis/custos/internal/semantic/infer"
)

// Options contains resolved configuration; adapters own flags and settings.
type Options struct {
	Root           string
	Paths, Exclude []string
	// Scope lists the project paths whose PHP files are indexed for symbols
	// besides the analyzed ones, so cross-file facts do not depend on which
	// files a run names. Empty means Paths.
	Scope []string
	// Cache is the directory keeping, between runs, the declarations of
	// files indexed but not analyzed (see SymbolCache) and the flow
	// summaries of the analyzed ones (see flowSteps); empty disables it.
	Cache string
	// FlowOrder summarizes the flow of the analyzed files in the order the
	// last run with the cache saved, and saves this run's (see
	// BuildFlowOrdered). A call chain composing without an order gives up
	// on may then settle, so results may differ from a run without it.
	FlowOrder bool
	// ParseOnce keeps the syntax tree of each analyzed file from indexing
	// until its analysis, so it is parsed once instead of once per phase
	// and per flow summary: faster, but every analyzed file's tree stays in
	// memory until the analysis reaches it.
	ParseOnce bool
	// RecordedKey identifies the baseline the engine's checks see (see
	// analysis.Engine.WithRecorded): the findings cache is kept for an
	// engine that sees one only when it is set.
	RecordedKey string
	Analysis    analysis.Config
	Parse       syntax.Options
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
	engine, sources, trees, found, saved := p.prepare(p.Options.ParseOnce, true)
	defer saved()
	results := runSources(engine, p.Files, sources, trees, p.Options.Parse, true, found)
	reported(found)
	// The cache only saves time: a failure to write it changes no result.
	_ = found.save()
	return results
}

// PrepareFixes computes changes without writing; the adapter owns writes.
func (p *Project) PrepareFixes() []PreparedFix {
	engine, sources, _, _, saved := p.prepare(false, false)
	defer saved()
	return PrepareFixes(engine, p.Files, sources, p.Options.Parse)
}

// prepare attaches the project index and flow snapshot the engine needs. It
// returns the project sources the index read (nil where unread), so analysis
// and fixes see the same bytes without reading each file again, and, with
// keepTrees, the syntax trees it parsed of them (see Options.ParseOnce). The
// flow steps are saved while the files are analyzed: saved waits until they
// are. With report, it also opens the findings cache of the run (nil when it
// has none, see findingsCache).
// Files that rules request by name, such as composer.json, are analyzed but
// never indexed or summarized as PHP.
func (p *Project) prepare(keepTrees, report bool) (engine *analysis.Engine, sources [][]byte, trees []*syntax.File, found *findingsCache, saved func()) {
	engine, saved = p.Engine, func() {}
	if !engine.NeedsIndex() {
		return engine, nil, nil, nil, saved
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
	inferOpt := infer.Options{Assertions: p.Options.Analysis.Flow.Assertions}
	var cache *SymbolCache
	if p.Options.Cache != "" {
		cache = OpenSymbolCache(p.Options.Cache, p.Options.Root, p.Options.Parse, inferOpt)
		cache.fingerprint(len(php))
	}
	indexed, skipped := IndexSourcesSkipping(p.Options.Root, php, scope, p.Options.Exclude)
	ix, read, parsed, unread := buildIndex(context.Background(), indexed, len(php), p.Options.Parse, inferOpt, cache, keepTrees)
	skipped.Files = append(skipped.Files, unread...)
	ix.SetWorld(ComposerWorld(p.Options.Root, scope, indexed, skipped, ix))
	ix.Freeze()
	engine = engine.WithIndex(ix)
	var snapshot *flow.Snapshot
	var options flow.Options
	if engine.NeedsFlow() {
		options = WithVendorFlow(p.Options.Root, ix, p.Options.Parse, p.Options.Analysis.Flow)
		var order map[string]int
		if p.Options.FlowOrder {
			order = cache.FlowOrder(php, options)
		}
		steps := cache.flowSteps(p.Options.Root, php, read, options)
		var earlier *FlowBuild
		if steps != nil {
			earlier = steps.earlier
		}
		build := buildFlow(context.Background(), php, read, p.Options.Parse, ix, options, order, earlier, steps != nil, parsed)
		if p.Options.FlowOrder && cache != nil {
			_ = cache.SaveFlowOrder(build.Order(), options)
		}
		// The cache only saves time: a failure to write it changes no result.
		done := make(chan struct{})
		go func() {
			defer close(done)
			_ = steps.save(build)
		}()
		saved = func() { <-done }
		preparedFlow(build)
		engine = engine.WithFlow(build.Snapshot)
		snapshot = build.Snapshot
	}
	if report {
		found = cache.findings(p.Files, findingsInputs{
			root: p.Options.Root, recorded: p.Options.RecordedKey, parseOnce: p.Options.ParseOnce,
			engine: engine, world: func() string { return worldDigest(p.Options.Root, indexed, skipped) }, php: php,
		}, snapshot, options)
	}
	sources = make([][]byte, len(p.Files))
	if keepTrees {
		trees = make([]*syntax.File, len(p.Files))
	}
	for j, i := range at {
		sources[i] = read[j]
		if keepTrees {
			trees[i] = parsed[j]
		}
	}
	return engine, sources, trees, found, saved
}

// AnalyzeBuffer analyzes supplied bytes without reading the filesystem.
func AnalyzeBuffer(engine fixing.Analyzer, path string, src []byte, opt syntax.Options) diagnostic.FileResult {
	return analyzeSource(engine, path, src, opt, nil)
}

// analyzeSource is AnalyzeBuffer analyzing file, when given, as the tree
// of src instead of parsing it.
func analyzeSource(engine fixing.Analyzer, path string, src []byte, opt syntax.Options, file *syntax.File) (result diagnostic.FileResult) {
	defer func() {
		if r := recover(); r != nil {
			result = internalFailure(path, src, r)
		}
	}()
	if file == nil {
		file = syntax.ParseBest(path, src, opt)
	}
	result = diagnostic.FileResult{Path: path, Src: src, Errors: file.Errors}
	if detailed, ok := engine.(interface {
		AnalyzeResult(*syntax.File) diagnostic.AnalysisResult
	}); ok {
		analysis := detailed.AnalyzeResult(file)
		result.Findings, result.Incomplete = analysis.Findings, analysis.Incomplete
	} else {
		result.Findings = engine.Analyze(file)
		if e, ok := file.RecoveryError(); ok {
			result.Incomplete = []diagnostic.AnalysisIssue{{Reason: "syntax-recovery", Span: e.Span}}
		}
	}
	return result
}

// preparedFlow sees each flow build prepare makes; tests replace it.
var preparedFlow = func(*FlowBuild) {}

// reported sees the findings cache of each AnalyzeReport once it analyzed
// the files (nil without one); tests replace it.
var reported = func(*findingsCache) {}
