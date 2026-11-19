// Package argumentcountmismatch implements the native ArgumentCountMismatch
// inspection.
package argumentcountmismatch

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }
