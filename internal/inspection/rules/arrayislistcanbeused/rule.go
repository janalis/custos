// Package arrayislistcanbeused implements the ArrayIsListCanBeUsed inspection.
package arrayislistcanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return arrayIsListCanBeUsed{} }
