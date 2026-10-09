// Package uselessunset implements the UselessUnset inspection.
package uselessunset

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return uselessUnset{} }
