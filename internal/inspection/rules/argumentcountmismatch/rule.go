// Package argumentcountmismatch implements the ArgumentCountMismatch inspection.
package argumentcountmismatch

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return rule{} }
