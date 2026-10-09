// Package offsetoperations implements the OffsetOperations inspection.
package offsetoperations

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return offsetOperations{} }
