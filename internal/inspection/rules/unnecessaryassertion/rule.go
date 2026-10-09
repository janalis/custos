// Package unnecessaryassertion implements the UnnecessaryAssertion inspection.
package unnecessaryassertion

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unnecessaryAssertion{} }
