// Package dateusage implements the DateUsage inspection.
package dateusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return dateUsage{} }
