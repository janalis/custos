// Package constantcanbeused implements the ConstantCanBeUsed inspection.
package constantcanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return constantCanBeUsed{} }
