// Package getclassusage implements the GetClassUsage inspection.
package getclassusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return getClassUsage{} }
