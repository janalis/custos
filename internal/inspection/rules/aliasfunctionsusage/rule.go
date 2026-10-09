// Package aliasfunctionsusage implements the AliasFunctionsUsage inspection.
package aliasfunctionsusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return aliasFunctionsUsage{} }
