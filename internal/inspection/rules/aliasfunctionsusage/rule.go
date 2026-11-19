// Package aliasfunctionsusage implements the AliasFunctionsUsage inspection.
package aliasfunctionsusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return aliasFunctionsUsage{} }
