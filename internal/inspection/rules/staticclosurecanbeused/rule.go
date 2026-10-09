// Package staticclosurecanbeused implements the StaticClosureCanBeUsed inspection.
package staticclosurecanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return staticClosureCanBeUsed{} }
