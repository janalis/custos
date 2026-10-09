// Package emptyclass implements the EmptyClass inspection.
package emptyclass

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return emptyClass{} }
