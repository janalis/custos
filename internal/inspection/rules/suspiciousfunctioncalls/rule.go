// Package suspiciousfunctioncalls implements the SuspiciousFunctionCalls inspection.
package suspiciousfunctioncalls

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return suspiciousFunctionCalls{} }
