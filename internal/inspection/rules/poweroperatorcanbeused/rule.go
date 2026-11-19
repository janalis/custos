// Package poweroperatorcanbeused implements the PowerOperatorCanBeUsed inspection.
package poweroperatorcanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return powerOperatorCanBeUsed{} }
