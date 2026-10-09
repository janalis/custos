package analysis

import (
	"custos/internal/diagnostic"
	"custos/internal/php/syntax"
	phpversion "custos/internal/php/version"
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

// ComparisonStyle is the preferred operand order in comparisons.
type ComparisonStyle uint8

const (
	StyleRegular ComparisonStyle = iota // $x === null
	StyleYoda                           // null === $x
)

// RuleConfig overrides catalogue defaults for one rule.
type RuleConfig struct {
	Enabled  *bool
	Severity diagnostic.Severity
	Options  map[string]any // typed per meta option (bool, int, string, []string)
}

// Config selects and configures rules.
type Config struct {
	PHP             phpversion.Version // target version for rule gating
	ComparisonStyle ComparisonStyle
	Rules           map[string]RuleConfig // keyed by rule ID
	// Only, when non-empty, enables exactly these rule IDs (overrides defaults).
	Only []string
	// EnableAll enables every registered rule, including disabled-by-default ones.
	EnableAll bool
}
