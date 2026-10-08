// Package analysis runs rules over parsed files: a single AST walk per file
// dispatches each node to the rules subscribed to its kind.
package analysis

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"custos/internal/index"
	"custos/internal/infer"
	"custos/internal/meta"
	"custos/internal/names"
	"custos/internal/phpver"
	"custos/internal/syntax"
)

// Rule is one inspection. Implementations must be stateless and safe for
// concurrent use: per-file state lives in Context.
type Rule interface {
	// ID is the custos rule ID (see internal/meta).
	ID() string
	// Kinds lists the node kinds Check wants to see.
	Kinds() []syntax.NodeKind
	// Check inspects one node.
	Check(ctx *Context, n syntax.Node)
}

// FileRule is implemented by rules that inspect a file as a whole (called
// once per file, before the walk).
type FileRule interface {
	CheckFile(ctx *Context)
}

// TextEdit replaces Span with NewText.
type TextEdit struct {
	Span    syntax.Span
	NewText string
}

// Fix is a lazily built quick-fix.
type Fix struct {
	Title string
	Edits func() []TextEdit
}

// Finding is one reported problem.
type Finding struct {
	Rule     string
	Severity meta.Severity
	Span     syntax.Span
	Message  string
	Fixes    []Fix
}

// ComparisonStyle is the preferred operand order in comparisons.
type ComparisonStyle uint8

const (
	StyleRegular ComparisonStyle = iota // $x === null
	StyleYoda                           // null === $x
)

// RuleConfig overrides catalogue defaults for one rule.
type RuleConfig struct {
	Enabled  *bool
	Severity meta.Severity
	Options  map[string]any // typed per meta option (bool, int, string, []string)
}

// Config selects and configures rules.
type Config struct {
	PHP             phpver.Version // target version for rule gating
	ComparisonStyle ComparisonStyle
	Rules           map[string]RuleConfig // keyed by rule ID
	// Only, when non-empty, enables exactly these rule IDs (overrides defaults).
	Only []string
	// EnableAll enables every registered rule, including disabled-by-default ones.
	EnableAll bool
}

type activeRule struct {
	rule     Rule
	meta     *meta.Rule
	severity meta.Severity
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
		cfg.PHP = phpver.Default
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
func (e *Engine) Analyze(f *syntax.File) []Finding {
	var disabled map[int]bool
	var internal []Finding
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
		internal = append(internal, Finding{
			Rule: "internal", Severity: meta.SeverityError,
			Message: fmt.Sprintf("custos rule %s crashed on this file and was skipped: %s", e.rules[crashed].meta.ID, msg),
		})
	}
}

// analyzeOnce runs the rules not in disabled. When a rule panics it returns
// that rule's index (else -1) and the panic message.
func (e *Engine) analyzeOnce(f *syntax.File, disabled map[int]bool) (out []Finding, crashed int, msg string) {
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
		out = append(out, Finding{
			Rule: "internal", Severity: meta.SeverityWarning,
			Message: fmt.Sprintf("more than %d findings in this file; the rest are not reported", MaxFindingsPerFile),
		})
	}
	return out, -1, ""
}

// Context is handed to rules; it is valid for one file.
type Context struct {
	File            *syntax.File
	Src             []byte
	PHP             phpver.Version
	ComparisonStyle ComparisonStyle

	engine   *Engine
	cur      *activeRule
	findings []Finding
	names    *names.Resolver
	index    *index.Index
	types    *infer.Env
	memo     map[string]any
	// truncated is set once MaxFindingsPerFile findings were collected.
	truncated bool
}

// Text returns the source text of n ("" for a nil node).
func (c *Context) Text(n syntax.Node) string {
	if n == nil || reflectNil(n) {
		return ""
	}
	return c.SpanText(n.Span())
}

// SpanText returns the source text of s ("" for an empty, inverted or
// out-of-range span, which error recovery or rule arithmetic can produce).
func (c *Context) SpanText(s syntax.Span) string {
	if s.End <= s.Start || int(s.End) > len(c.Src) {
		return ""
	}
	return string(c.Src[s.Start:s.End])
}

// Report records a finding at span with the rule's configured severity.
func (c *Context) Report(span syntax.Span, msg string, fixes ...Fix) {
	c.ReportSeverity(span, c.cur.severity, msg, fixes...)
}

// ReportNode records a finding covering node n.
func (c *Context) ReportNode(n syntax.Node, msg string, fixes ...Fix) {
	c.Report(n.Span(), msg, fixes...)
}

// ReportSeverity records a finding with an explicit severity (for rules that
// report several problem classes).
func (c *Context) ReportSeverity(span syntax.Span, sev meta.Severity, msg string, fixes ...Fix) {
	if len(c.findings) >= MaxFindingsPerFile {
		c.truncated = true
		return
	}
	c.findings = append(c.findings, Finding{Rule: c.cur.meta.ID, Severity: sev, Span: span, Message: msg, Fixes: fixes})
}

// MaxFindingsPerFile caps the findings collected for one file. Generated,
// minified or hostile files can otherwise yield millions of findings
// (gigabytes of memory and output, e.g. 20 MB of `;`); beyond the cap the
// rest are dropped and one "internal" note says so.
const MaxFindingsPerFile = 10000

// RuleID returns the ID of the rule currently running.
func (c *Context) RuleID() string { return c.cur.meta.ID }

// OptionSet reports whether the configuration sets option name of the
// current rule explicitly (false when the catalogue default applies).
func (c *Context) OptionSet(name string) bool { return c.cur.explicit[name] }

// Bool returns a boolean option of the current rule.
func (c *Context) Bool(name string) bool {
	switch v := c.cur.options[name].(type) {
	case bool:
		return v
	case string:
		b, _ := strconv.ParseBool(v)
		return b
	}
	return false
}

// Int returns an integer option of the current rule.
func (c *Context) Int(name string) int {
	switch v := c.cur.options[name].(type) {
	case int:
		return v
	case float64:
		return int(v)
	case string:
		n, _ := strconv.Atoi(v)
		return n
	}
	return 0
}

// String returns a string/enum option of the current rule.
func (c *Context) String(name string) string {
	if v, ok := c.cur.options[name].(string); ok {
		return v
	}
	return ""
}

// List returns a list option of the current rule.
func (c *Context) List(name string) []string {
	switch v := c.cur.options[name].(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, x := range v {
			out = append(out, fmt.Sprint(x))
		}
		return out
	}
	return nil
}

// IsTestFile reports whether the file looks like a test (IsTestPath).
func (c *Context) IsTestFile() bool { return IsTestPath(c.File.Path) }

// IsTestPath reports whether a file path denotes a test context: it ends
// with Test.php, Spec.php or .phpt, or contains a /Fixtures/ directory
// (case-sensitive; `\` separators accepted).
func IsTestPath(path string) bool {
	p := strings.ReplaceAll(path, `\`, "/")
	return strings.HasSuffix(p, "Test.php") || strings.HasSuffix(p, "Spec.php") ||
		strings.HasSuffix(p, ".phpt") || strings.Contains(p, "/Fixtures/")
}

func sortFindings(fs []Finding) {
	// insertion sort keeps it allocation-free; findings per file are few.
	for i := 1; i < len(fs); i++ {
		for j := i; j > 0 && less(fs[j], fs[j-1]); j-- {
			fs[j], fs[j-1] = fs[j-1], fs[j]
		}
	}
}

func less(a, b Finding) bool {
	if a.Span.Start != b.Span.Start {
		return a.Span.Start < b.Span.Start
	}
	if a.Span.End != b.Span.End {
		return a.Span.End < b.Span.End
	}
	return a.Rule < b.Rule
}

// reflectNil detects typed-nil nodes (e.g. a nil *Literal stored in an Expr).
func reflectNil(n syntax.Node) bool {
	v := reflect.ValueOf(n)
	return v.Kind() == reflect.Pointer && v.IsNil()
}
