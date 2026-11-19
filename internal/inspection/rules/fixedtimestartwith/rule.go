// Package fixedtimestartwith implements the FixedTimeStartWith inspection.
package fixedtimestartwith

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return fixedTimeStartWith{} }
