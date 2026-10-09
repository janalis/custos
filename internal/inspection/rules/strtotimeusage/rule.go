// Package strtotimeusage implements the StrtotimeUsage inspection.
package strtotimeusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return strtotimeUsage{} }
