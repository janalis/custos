// Package unnecessarycasting implements the UnnecessaryCasting inspection.
package unnecessarycasting

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unnecessaryCasting{} }
