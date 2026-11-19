// Package unnecessaryfinalmodifier implements the UnnecessaryFinalModifier inspection.
package unnecessaryfinalmodifier

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unnecessaryFinalModifier{} }
