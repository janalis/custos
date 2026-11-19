// Package variablefunctionsusage implements the VariableFunctionsUsage inspection.
package variablefunctionsusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return variableFunctionsUsage{} }
