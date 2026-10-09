package analysis

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"custos/internal/diagnostic"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
	"custos/internal/semantic/index"
	"custos/internal/semantic/infer"
	"custos/internal/semantic/names"
)

// Context is handed to rules; it is valid for one file.
type Context struct {
	File            *syntax.File
	Src             []byte
	PHP             phpversion.Version
	ComparisonStyle ComparisonStyle

	engine   *Engine
	cur      *activeRule
	findings []diagnostic.Finding
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
func (c *Context) Report(span syntax.Span, msg string, fixes ...diagnostic.Fix) {
	c.ReportSeverity(span, c.cur.severity, msg, fixes...)
}

// ReportNode records a finding covering node n.
func (c *Context) ReportNode(n syntax.Node, msg string, fixes ...diagnostic.Fix) {
	c.Report(n.Span(), msg, fixes...)
}

// ReportSeverity records a finding with an explicit severity (for rules that
// report several problem classes).
func (c *Context) ReportSeverity(span syntax.Span, sev diagnostic.Severity, msg string, fixes ...diagnostic.Fix) {
	if len(c.findings) >= MaxFindingsPerFile {
		c.truncated = true
		return
	}
	c.findings = append(c.findings, diagnostic.Finding{Rule: c.cur.meta.ID, Severity: sev, Span: span, Message: msg, Fixes: fixes})
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

// reflectNil detects typed-nil nodes (e.g. a nil *Literal stored in an Expr).
func reflectNil(n syntax.Node) bool {
	v := reflect.ValueOf(n)
	return v.Kind() == reflect.Pointer && v.IsNil()
}
