// Package loopwhichdoesnotloop implements the LoopWhichDoesNotLoop inspection.
package loopwhichdoesnotloop

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return loopWhichDoesNotLoop{} }
