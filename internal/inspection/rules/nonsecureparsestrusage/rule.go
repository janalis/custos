// Package nonsecureparsestrusage implements the NonSecureParseStrUsage inspection.
package nonsecureparsestrusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return nonSecureParseStrUsage{} }
