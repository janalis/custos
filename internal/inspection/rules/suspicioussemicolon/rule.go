// Package suspicioussemicolon implements the SuspiciousSemicolon inspection.
package suspicioussemicolon

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return suspiciousSemicolon{} }
