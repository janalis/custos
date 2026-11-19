// Package elvisoperatorcanbeused implements the ElvisOperatorCanBeUsed inspection.
package elvisoperatorcanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return elvisOperatorCanBeUsed{} }
