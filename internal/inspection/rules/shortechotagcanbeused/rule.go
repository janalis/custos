// Package shortechotagcanbeused implements the ShortEchoTagCanBeUsed inspection.
package shortechotagcanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return shortEchoTagCanBeUsed{} }
