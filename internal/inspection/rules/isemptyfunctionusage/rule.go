// Package isemptyfunctionusage implements the IsEmptyFunctionUsage inspection.
package isemptyfunctionusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return isEmptyFunctionUsage{} }
