// Package compactarguments implements the CompactArguments inspection.
package compactarguments

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return compactArguments{} }
