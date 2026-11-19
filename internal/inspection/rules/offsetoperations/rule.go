// Package offsetoperations implements the OffsetOperations inspection.
package offsetoperations

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return offsetOperations{} }
