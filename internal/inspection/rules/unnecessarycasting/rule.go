// Package unnecessarycasting implements the UnnecessaryCasting inspection.
package unnecessarycasting

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unnecessaryCasting{} }
