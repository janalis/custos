package analysis

import (
	"github.com/janalis/custos/internal/diagnostic"
	"github.com/janalis/custos/internal/php/syntax"
	phpversion "github.com/janalis/custos/internal/php/version"
	"github.com/janalis/custos/internal/semantic/flow"
)

// Rule is one inspection. Implementations must be stateless and safe for
// concurrent use: per-file state lives in Context.
type Rule interface {
	// ID is the custos rule ID (see internal/inspection/meta).
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

// FormatRule marks rules that lay out code as the formatter does: they
// read the settings Config.Format gives each file.
type FormatRule interface{ Format() }

// ComparisonStyle is the preferred operand order in comparisons.
type ComparisonStyle uint8

const (
	StyleRegular ComparisonStyle = iota // $x === null
	StyleYoda                           // null === $x
)

// TypeCheck is how much type checking the typeCheck setting turns on.
type TypeCheck uint8

const (
	TypeCheckOff        TypeCheck = iota // only the rules enabled on their own
	TypeCheckProven                      // mismatches the code alone proves
	TypeCheckDocumented                  // also those against PHPDoc contracts
)

// Includes reports whether the level enables the rules of group, a rule's
// meta.Rule.TypeCheck.
func (l TypeCheck) Includes(group string) bool {
	switch group {
	case "proven":
		return l >= TypeCheckProven
	case "documented":
		return l >= TypeCheckDocumented
	}
	return false
}

// RuleConfig overrides catalogue defaults for one rule.
type RuleConfig struct {
	Enabled  *bool
	Severity diagnostic.Severity
	Options  map[string]any // typed per meta option (bool, int, string, []string)
}

// Config selects and configures rules.
type Config struct {
	Flow            flow.Options
	PHP             phpversion.Version // target version for rule gating
	ComparisonStyle ComparisonStyle
	TypeCheck       TypeCheck             // rule groups enabled together
	Rules           map[string]RuleConfig // keyed by rule ID
	// Only, when non-empty, enables exactly these rule IDs (overrides defaults).
	Only []string
	// EnableAll enables every registered rule, including disabled-by-default
	// ones, except those the rule config explicitly disables.
	EnableAll bool
	// ProjectPath maps a file path to its path relative to its project
	// root, with forward slashes, for rules matching project path globs.
	// Adapters supply it, as they own filesystem access; nil keeps paths
	// as given.
	ProjectPath func(path string) string
	// Format returns the formatter settings of the file at path for the
	// rules that format code (see FormatRule): an opaque, comparable value
	// those rules know (nil for the default settings), and false when the
	// file is not to be formatted. Adapters supply it, as they own the
	// project configuration; nil formats every file with the default
	// settings.
	Format func(path string) (settings any, ok bool)
	// LabelRules runs each rule's checks under a "rule" profiler label
	// carrying its ID, so a CPU profile splits the time rules take by rule.
	LabelRules bool
}
