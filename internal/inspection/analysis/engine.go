package analysis

import (
	"fmt"

	"custos/internal/diagnostic"
	"custos/internal/inspection/meta"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
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
		switch {
		case len(only) > 0:
			enabled = only[m.ID]
		case cfg.EnableAll:
			enabled = true
		case rc.Enabled != nil:
			enabled = *rc.Enabled
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
	var disabled map[int]bool
	var internal []diagnostic.Finding
	// Terminates: each crash disables a distinct rule, and a run with every
	// rule disabled cannot crash.
	for {
		findings, crashed, msg := e.analyzeOnce(f, disabled)
		if crashed < 0 {
			return append(internal, findings...)
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

// analyzeOnce runs the rules not in disabled. When a rule panics it returns
// that rule's index (else -1) and the panic message.
func (e *Engine) analyzeOnce(f *syntax.File, disabled map[int]bool) (out []diagnostic.Finding, crashed int, msg string) {
	ctx := &Context{File: f, Src: f.Src, PHP: e.cfg.PHP, ComparisonStyle: e.cfg.ComparisonStyle, engine: e}
	cur := -1
	defer func() {
		if r := recover(); r != nil {
			out, crashed, msg = nil, cur, fmt.Sprint(r)
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
		e.rules[i].rule.(FileRule).CheckFile(ctx)
	}
	if len(e.rules) > len(e.files) || len(e.files) == 0 {
		syntax.InspectFile(f, func(n syntax.Node) bool {
			for _, i := range e.byKind[n.Kind()] {
				if disabled != nil && disabled[int(i)] {
					continue
				}
				cur = int(i)
				ctx.cur = &e.rules[i]
				e.rules[i].rule.Check(ctx, n)
			}
			return true
		})
	}
	cur = -1
	if len(ctx.findings) == 0 {
		return nil, -1, ""
	}
	sup := newSuppressions(f)
	out = ctx.findings[:0]
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
	return out, -1, ""
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
