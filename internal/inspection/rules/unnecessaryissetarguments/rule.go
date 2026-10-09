// Package unnecessaryissetarguments implements the UnnecessaryIssetArguments inspection.
package unnecessaryissetarguments

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unnecessaryIssetArguments{} }
