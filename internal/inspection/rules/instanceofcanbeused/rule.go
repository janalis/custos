// Package instanceofcanbeused implements the InstanceofCanBeUsed inspection.
package instanceofcanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return instanceofCanBeUsed{} }
