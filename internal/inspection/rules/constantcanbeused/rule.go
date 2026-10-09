// Package constantcanbeused implements the ConstantCanBeUsed inspection.
package constantcanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return constantCanBeUsed{} }
