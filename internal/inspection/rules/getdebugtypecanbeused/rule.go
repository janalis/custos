// Package getdebugtypecanbeused implements the GetDebugTypeCanBeUsed inspection.
package getdebugtypecanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return getDebugTypeCanBeUsed{} }
