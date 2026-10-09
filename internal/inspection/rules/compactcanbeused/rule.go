// Package compactcanbeused implements the CompactCanBeUsed inspection.
package compactcanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return compactCanBeUsed{} }
