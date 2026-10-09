// Package arrayislistcanbeused implements the ArrayIsListCanBeUsed inspection.
package arrayislistcanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return arrayIsListCanBeUsed{} }
