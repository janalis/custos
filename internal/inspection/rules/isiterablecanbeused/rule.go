// Package isiterablecanbeused implements the IsIterableCanBeUsed inspection.
package isiterablecanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return isIterableCanBeUsed{} }
