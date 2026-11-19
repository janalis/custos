// Package foreachinvariants implements the ForeachInvariants inspection.
package foreachinvariants

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return foreachInvariants{} }
