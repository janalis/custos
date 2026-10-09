// Package foreachinvariants implements the ForeachInvariants inspection.
package foreachinvariants

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return foreachInvariants{} }
