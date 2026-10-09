// Package pregquoteusage implements the PregQuoteUsage inspection.
package pregquoteusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return pregQuoteUsage{} }
