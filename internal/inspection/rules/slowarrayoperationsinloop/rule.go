// Package slowarrayoperationsinloop implements the SlowArrayOperationsInLoop inspection.
package slowarrayoperationsinloop

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return slowArrayOperationsInLoop{} }
