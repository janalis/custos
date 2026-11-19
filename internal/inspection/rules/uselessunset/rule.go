// Package uselessunset implements the UselessUnset inspection.
package uselessunset

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return uselessUnset{} }
