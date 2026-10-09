// Package switchcontinuationinloop implements the SwitchContinuationInLoop inspection.
package switchcontinuationinloop

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return switchContinuationInLoop{} }
