// Package suspiciousloop implements the SuspiciousLoop inspection.
package suspiciousloop

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return suspiciousLoop{} }
