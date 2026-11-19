// Package strtotimeusage implements the StrtotimeUsage inspection.
package strtotimeusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return strtotimeUsage{} }
