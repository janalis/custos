// Package arrayuniquecanbeused implements the ArrayUniqueCanBeUsed inspection.
package arrayuniquecanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return arrayUniqueCanBeUsed{} }
