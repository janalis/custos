// Package typescastingcanbeused implements the TypesCastingCanBeUsed inspection.
package typescastingcanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return typesCastingCanBeUsed{} }
