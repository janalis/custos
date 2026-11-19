// Package compactcanbeused implements the CompactCanBeUsed inspection.
package compactcanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return compactCanBeUsed{} }
