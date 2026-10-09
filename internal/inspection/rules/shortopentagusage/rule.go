// Package shortopentagusage implements the ShortOpenTagUsage inspection.
package shortopentagusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return shortOpenTagUsage{} }
