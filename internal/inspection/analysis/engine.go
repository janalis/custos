package analysis

import (
	"fmt"
	"maps"
	"slices"

	"custos/internal/diagnostic"
	"custos/internal/inspection/meta"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/flow"
	"custos/internal/semantic/index"
)

type activeRule struct {
	rule     Rule
	meta     *meta.Rule
	severity diagnostic.Severity
	options  map[string]any
	explicit map[string]bool // options set by the configuration (not defaults)
}

// Engine is a configured set of rules. It is safe for concurrent use.
type Engine struct {
	cfg    Config
	rules  []activeRule
	byKind [syntax.NumNodeKinds][]int32
	files  []int32 // indices of FileRule implementations
	index  *index.Index
	flow   *flow.Snapshot
}

// NewEngine builds an engine from the registered rules and cfg.
func NewEngine(registered []Rule, cfg Config) (*Engine, error) {
	if cfg.PHP == 0 {
		cfg.PHP = phpversion.Default
	}
	only := map[string]bool{}
	for _, id := range cfg.Only {
		only[id] = true
	}
	e := &Engine{cfg: cfg}
	for _, r := range registered {
		m, ok := meta.Lookup(r.ID())
		if !ok {
			return nil, fmt.Errorf("analysis: rule %s missing from catalogue", r.ID())
		}
		rc := cfg.Rules[m.ID]
		enabled := m.EnabledByDefault
		// Naming rules beats the config; the config's explicit choice
		// beats enabling everything.
		switch {
		case len(only) > 0:
			enabled = only[m.ID]
		case rc.Enabled != nil:
			enabled = *rc.Enabled
		case cfg.EnableAll:
			enabled = true
		}
		if !enabled {
			continue
		}
		ar := activeRule{rule: r, meta: m, severity: m.Severity, options: map[string]any{}}
		if rc.Severity != "" {
			ar.severity = rc.Severity
		}
		for _, o := range m.Options {
			ar.options[o.Name] = o.Default
		}
		for k, v := range rc.Options {
			ar.options[k] = v
			if ar.explicit == nil {
				ar.explicit = map[string]bool{}
			}
			ar.explicit[k] = true
		}
		idx := int32(len(e.rules))
		e.rules = append(e.rules, ar)
		for _, k := range r.Kinds() {
			e.byKind[k] = append(e.byKind[k], idx)
		}
		if _, ok := r.(FileRule); ok {
			e.files = append(e.files, idx)
		}
	}
	return e, nil
}

// Rules returns the IDs of the enabled rules.
func (e *Engine) Rules() []string {
	ids := make([]string, len(e.rules))
	for i, r := range e.rules {
		ids[i] = r.meta.ID
	}
	return ids
}

// Config returns the engine configuration.
func (e *Engine) Config() Config { return e.cfg }

// Analyze runs all enabled rules on f and returns unsuppressed findings in
// source order. A rule that panics is reported as an "internal" error
// finding and disabled for this file; the remaining rules still run.
func (e *Engine) Analyze(f *syntax.File) []diagnostic.Finding {
	return e.AnalyzeResult(f).Findings
}

// AnalyzeResult also reports incomplete proofs encountered by enabled checks.
func (e *Engine) AnalyzeResult(f *syntax.File) diagnostic.AnalysisResult {
	var disabled map[int]bool
	var internal []diagnostic.Finding
	failed := map[string]envFailure{}
	// Terminates: each crash disables a distinct rule or records the first
	// failure of shared data (later runs isolate checks from any), and a
	// run with every rule disabled cannot crash.
	for {
		result, crashed, msg := e.analyzeOnce(f, disabled, failed)
		if crashed == envCrashed {
			continue
		}
		if crashed < 0 {
			internal = append(internal, envFindings(failed)...)
			result.Findings = append(internal, result.Findings...)
			reason := "rule-failure"
			if len(failed) > 0 {
				reason = "internal-failure"
			}
			if len(internal) > 0 {
				result.Incomplete = append(result.Incomplete, diagnostic.AnalysisIssue{Reason: reason, Span: f.Span()})
			}
			return result
		}
		if disabled == nil {
			disabled = map[int]bool{}
		}
		disabled[crashed] = true
		internal = append(internal, diagnostic.Finding{
			Rule: "internal", Severity: diagnostic.SeverityError,
			Message: fmt.Sprintf("custos rule %s crashed on this file and was skipped: %s", e.rules[crashed].meta.ID, msg),
		})
	}
}

// envCrashed is analyzeOnce's crashed index for a failure of shared data.
const envCrashed = -2

// envFindings reports each failed build of shared data once (data built
// from failed data names the failure it inherited).
func envFindings(failed map[string]envFailure) []diagnostic.Finding {
	var out []diagnostic.Finding
	seen := map[string]bool{}
	for _, what := range slices.Sorted(maps.Keys(failed)) {
		f := failed[what]
		if seen[f.what] {
			continue
		}
		seen[f.what] = true
		out = append(out, diagnostic.Finding{
			Rule: "internal", Severity: diagnostic.SeverityError,
			Message: fmt.Sprintf("custos could not complete the %s of this file and skipped the checks needing it: %v", f.what, f.value),
		})
	}
	return out
}

// analyzeOnce runs the rules not in disabled. When a rule panics it returns
// that rule's index (else -1) and the panic message; when building data
// the rules share fails, envCrashed. Once some failed
// (failed is not empty), each check runs isolated: one needing that data
// is skipped and the others go on.
func (e *Engine) analyzeOnce(f *syntax.File, disabled map[int]bool, failed map[string]envFailure) (result diagnostic.AnalysisResult, crashed int, msg string) {
	ctx := &Context{File: f, Src: f.Src, PHP: e.cfg.PHP, ComparisonStyle: e.cfg.ComparisonStyle, engine: e, failed: failed}
	isolated := len(failed) > 0
	cur := -1
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(envFailure); ok {
				result, crashed, msg = diagnostic.AnalysisResult{}, envCrashed, ""
				return
			}
			result, crashed, msg = diagnostic.AnalysisResult{}, cur, fmt.Sprint(r)
			if cur < 0 {
				panic(r) // not a rule failure
			}
		}
	}()
	for _, i := range e.files {
		if disabled[int(i)] {
			continue
		}
		cur = int(i)
		ctx.cur = &e.rules[i]
		rule := e.rules[i].rule.(FileRule)
		if isolated {
			isolate(func() { rule.CheckFile(ctx) })
		} else {
			rule.CheckFile(ctx)
		}
	}
	if len(e.rules) > len(e.files) || len(e.files) == 0 {
		syntax.InspectFile(f, func(n syntax.Node) bool {
			for _, i := range e.byKind[n.Kind()] {
				if disabled != nil && disabled[int(i)] {
					continue
				}
				cur = int(i)
				ctx.cur = &e.rules[i]
				if isolated {
					isolate(func() { e.rules[i].rule.Check(ctx, n) })
				} else {
					e.rules[i].rule.Check(ctx, n)
				}
			}
			return true
		})
	}
	cur = -1
	if len(ctx.findings) == 0 {
		return diagnostic.AnalysisResult{Incomplete: ctx.incomplete()}, -1, ""
	}
	sup := newSuppressions(f)
	out := ctx.findings[:0]
	for _, fd := range ctx.findings {
		if !sup.suppressed(fd) {
			out = append(out, fd)
		}
	}
	sortFindings(out)
	if ctx.truncated {
		out = append(out, diagnostic.Finding{
			Rule: "internal", Severity: diagnostic.SeverityWarning,
			Message: fmt.Sprintf("more than %d findings in this file; the rest are not reported", MaxFindingsPerFile),
		})
	}
	return diagnostic.AnalysisResult{Findings: out, Incomplete: ctx.incomplete()}, -1, ""
}

// isolate runs check, skipping it when it needs shared data that failed to
// build; any other panic goes on to the rule's crash handling.
func isolate(check func()) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(envFailure); !ok {
				panic(r)
			}
		}
	}()
	check()
}

func sortFindings(fs []diagnostic.Finding) {
	// insertion sort keeps it allocation-free; findings per file are few.
	for i := 1; i < len(fs); i++ {
		for j := i; j > 0 && less(fs[j], fs[j-1]); j-- {
			fs[j], fs[j-1] = fs[j-1], fs[j]
		}
	}
}

func less(a, b diagnostic.Finding) bool {
	if a.Span.Start != b.Span.Start {
		return a.Span.Start < b.Span.Start
	}
	if a.Span.End != b.Span.End {
		return a.Span.End < b.Span.End
	}
	return a.Rule < b.Rule
}

func (c *Context) incomplete() []diagnostic.AnalysisIssue {
	var issues []diagnostic.AnalysisIssue
	if len(c.File.Errors) > 0 {
		issues = append(issues, diagnostic.AnalysisIssue{Reason: "syntax-recovery", Span: c.File.Errors[0].Span})
	}
	if c.types != nil {
		for _, gap := range c.types.Gaps() {
			issues = append(issues, diagnostic.AnalysisIssue{Reason: gap.Reason, Span: gap.Span})
		}
	}
	if c.flow != nil {
		for _, gap := range c.flow.Gaps() {
			issues = append(issues, diagnostic.AnalysisIssue{Reason: gap.Reason, Span: gap.Span})
		}
	}
	issues = append(issues, c.gaps...)
	if c.truncated {
		issues = append(issues, diagnostic.AnalysisIssue{Reason: "finding-limit", Span: c.File.Span()})
	}
	return issues
}
