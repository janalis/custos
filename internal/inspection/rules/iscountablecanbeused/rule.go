// Package iscountablecanbeused implements the IsCountableCanBeUsed inspection.
package iscountablecanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return isCountableCanBeUsed{} }
