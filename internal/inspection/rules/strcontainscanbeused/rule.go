// Package strcontainscanbeused implements the StrContainsCanBeUsed inspection.
package strcontainscanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return strContainsCanBeUsed{} }
