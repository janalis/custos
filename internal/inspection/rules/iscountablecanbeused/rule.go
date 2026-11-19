// Package iscountablecanbeused implements the IsCountableCanBeUsed inspection.
package iscountablecanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return isCountableCanBeUsed{} }
