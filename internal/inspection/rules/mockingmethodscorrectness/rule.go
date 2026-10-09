// Package mockingmethodscorrectness implements the MockingMethodsCorrectness inspection.
package mockingmethodscorrectness

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return mockingMethodsCorrectness{} }
