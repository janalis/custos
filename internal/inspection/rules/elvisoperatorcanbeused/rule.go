// Package elvisoperatorcanbeused implements the ElvisOperatorCanBeUsed inspection.
package elvisoperatorcanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return elvisOperatorCanBeUsed{} }
