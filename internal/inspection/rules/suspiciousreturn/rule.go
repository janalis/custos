// Package suspiciousreturn implements the SuspiciousReturn inspection.
package suspiciousreturn

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return suspiciousReturn{} }
