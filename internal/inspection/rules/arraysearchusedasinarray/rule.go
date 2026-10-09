// Package arraysearchusedasinarray implements the ArraySearchUsedAsInArray inspection.
package arraysearchusedasinarray

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return arraySearchUsedAsInArray{} }
