// Package custos is the Go API for rules written in Go. A program that
// imports it and calls Main is a complete custos binary (analyse, fix,
// explain, rules, the language server and the MCP server) that also runs
// its own rules:
//
//	package main
//
//	import (
//		"github.com/janalis/custos/pkg/custos"
//		"github.com/janalis/custos/pkg/custos/ast"
//	)
//
//	func main() {
//		noEval := custos.MustRule(custos.Info{ID: "NoEval", Summary: "eval runs arbitrary code."},
//			[]ast.NodeKind{ast.KEval},
//			func(ctx *custos.Context, n ast.Node) { ctx.ReportNode(n, "Remove eval") })
//		custos.Main("my-custos", noEval)
//	}
//
// Rules can also be declared as patterns, as in custos.json (Pattern).
//
// The API is unstable until custos 1.0: it exposes the engine's own types.
package custos

import (
	"errors"
	"io"
	"os"
	"slices"

	"github.com/janalis/custos/internal/cli"
	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/inspection/analysis"
	"github.com/janalis/custos/internal/inspection/meta"
	"github.com/janalis/custos/internal/inspection/pattern"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/pkg/custos/ast"
)

// The engine's contracts.
type (
	// Rule is one inspection. Implementations must be stateless and safe
	// for concurrent use; per-file state lives in Context.
	Rule = analysis.Rule
	// FileRule is implemented by rules that inspect a file as a whole.
	FileRule = analysis.FileRule
	// SemanticRule marks rules that need the project's declarations
	// (other files, dependencies) indexed before analysis.
	SemanticRule = analysis.SemanticRule
	// Described is implemented by rules carrying their own facts, as every
	// rule outside the built-in catalogue must.
	Described = analysis.Described
	// Context gives a rule the file under analysis, name resolution, type
	// inference and reporting.
	Context = analysis.Context
	// Facts are a rule's catalogue facts.
	Facts = meta.Rule
	// Description is a rule's documentation.
	Description = meta.Description
	// Severity is a finding's level.
	Severity = diagnostic.Severity
	// Fix is a quick-fix: a title and lazily built edits.
	Fix = diagnostic.Fix
	// TextEdit replaces a byte range of the source.
	TextEdit = diagnostic.TextEdit
	// PatternDefinition declares a pattern rule, as in custos.json.
	PatternDefinition = pattern.Definition
	// Where constrains a pattern metavariable.
	Where = pattern.Where
	// Patterns is one pattern or a list of them.
	Patterns = pattern.Patterns
)

// Severities.
const (
	SeverityError   = diagnostic.SeverityError
	SeverityWarning = diagnostic.SeverityWarning
	SeverityInfo    = diagnostic.SeverityInfo
)

// Group is the rule group of every rule outside the built-in catalogue.
const Group = pattern.Group

// Info describes a rule written in Go.
type Info struct {
	ID       string   // an uppercase letter, then letters and digits
	Severity Severity // default SeverityWarning
	Disabled bool     // off unless enabled in custos.json or run with --rule
	Summary  string   // shown by custos explain and in SARIF output
	HasFix   bool     // the rule offers quick-fixes
	PHPMin   string   // minimum target PHP version, documentation only
	// PartialCalls has the rule also check calls that only build a closure:
	// first-class callables and partial applications (`f(...)`, `f(?, 1)`,
	// PHP 8.6), whose ArgList.Partial() is true and whose arguments may be
	// ArgPlaceholder or VariadicPlaceholder nodes. Without it, such calls
	// never reach the rule.
	PartialCalls bool
}

// NewRule returns a rule calling check for each node of the given kinds.
func NewRule(info Info, kinds []ast.NodeKind, check func(ctx *Context, n ast.Node)) (Rule, error) {
	if err := pattern.CheckID(info.ID); err != nil {
		return nil, err
	}
	switch info.Severity {
	case "":
		info.Severity = SeverityWarning
	case SeverityError, SeverityWarning, SeverityInfo:
	default:
		return nil, errors.New("custos: invalid severity " + string(info.Severity))
	}
	if info.PHPMin != "" {
		if _, err := phpversion.Parse(info.PHPMin); err != nil {
			return nil, err
		}
	}
	if len(kinds) == 0 || check == nil {
		return nil, errors.New("custos: a rule needs node kinds and a check")
	}
	r := &goRule{info: info, kinds: slices.Clone(kinds), check: check}
	if info.PartialCalls {
		return &partialGoRule{r}, nil
	}
	return r, nil
}

// MustRule is NewRule for rules declared at startup: it panics on an
// invalid rule.
func MustRule(info Info, kinds []ast.NodeKind, check func(ctx *Context, n ast.Node)) Rule {
	r, err := NewRule(info, kinds, check)
	if err != nil {
		panic(err)
	}
	return r
}

// Pattern compiles a rule declared as a code pattern (see the custom rules
// guide).
func Pattern(d PatternDefinition) (Rule, error) { return pattern.Compile(d) }

// Run runs the custos command line with args (without the program name)
// and rules added to the built-in ones, and returns its exit status.
func Run(version string, args []string, stdin io.Reader, stdout, stderr io.Writer, rules ...Rule) int {
	return cli.RunWith(version, args, stdin, stdout, stderr, rules)
}

// exit is os.Exit; tests replace it.
var exit = os.Exit

// Main runs the custos command line on the process arguments with rules
// added to the built-in ones, and exits with its status.
func Main(version string, rules ...Rule) {
	exit(Run(version, os.Args[1:], os.Stdin, os.Stdout, os.Stderr, rules...))
}

// goRule is a rule written in Go with NewRule.
type goRule struct {
	info  Info
	kinds []ast.NodeKind
	check func(*Context, ast.Node)
}

func (r *goRule) ID() string                     { return r.info.ID }
func (r *goRule) Kinds() []ast.NodeKind          { return r.kinds }
func (r *goRule) Check(ctx *Context, n ast.Node) { r.check(ctx, n) }

// partialGoRule is a goRule with Info.PartialCalls.
type partialGoRule struct{ *goRule }

var _ analysis.PartialCallRule = partialGoRule{}

// PartialCalls marks the rule for the engine (see Info.PartialCalls).
func (partialGoRule) PartialCalls() {}

// Facts returns the rule's facts.
func (r *goRule) Facts() Facts {
	return Facts{Native: true, ID: r.info.ID, Group: Group, Severity: r.info.Severity, EnabledByDefault: !r.info.Disabled, HasFix: r.info.HasFix}
}

// Description returns the rule's documentation.
func (r *goRule) Description() Description {
	return Description{Summary: r.info.Summary, Fix: r.info.HasFix, PHPMin: r.info.PHPMin}
}
