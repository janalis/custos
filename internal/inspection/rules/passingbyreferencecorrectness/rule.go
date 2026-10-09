// Package passingbyreferencecorrectness implements the PassingByReferenceCorrectness inspection.
package passingbyreferencecorrectness

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return passingByReferenceCorrectness{} }
