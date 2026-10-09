// Package arrayuniquecanbeused implements the ArrayUniqueCanBeUsed inspection.
package arrayuniquecanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return arrayUniqueCanBeUsed{} }
