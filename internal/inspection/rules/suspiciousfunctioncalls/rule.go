// Package suspiciousfunctioncalls implements the SuspiciousFunctionCalls inspection.
package suspiciousfunctioncalls

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return suspiciousFunctionCalls{} }
