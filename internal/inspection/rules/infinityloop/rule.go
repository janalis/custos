// Package infinityloop implements the InfinityLoop inspection.
package infinityloop

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return infinityLoop{} }
