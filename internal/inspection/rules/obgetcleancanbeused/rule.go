// Package obgetcleancanbeused implements the ObGetCleanCanBeUsed inspection.
package obgetcleancanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return obGetCleanCanBeUsed{} }
