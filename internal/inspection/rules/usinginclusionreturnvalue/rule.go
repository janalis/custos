// Package usinginclusionreturnvalue implements the UsingInclusionReturnValue inspection.
package usinginclusionreturnvalue

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return usingInclusionReturnValue{} }
