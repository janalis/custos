// Package slowarrayoperationsinloop implements the SlowArrayOperationsInLoop inspection.
package slowarrayoperationsinloop

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return slowArrayOperationsInLoop{} }
