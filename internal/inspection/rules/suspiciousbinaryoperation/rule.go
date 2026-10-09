// Package suspiciousbinaryoperation implements the SuspiciousBinaryOperation inspection.
package suspiciousbinaryoperation

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return suspiciousBinaryOperation{} }
