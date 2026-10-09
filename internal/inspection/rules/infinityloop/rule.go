// Package infinityloop implements the InfinityLoop inspection.
package infinityloop

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return infinityLoop{} }
