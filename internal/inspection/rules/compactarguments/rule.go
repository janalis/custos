// Package compactarguments implements the CompactArguments inspection.
package compactarguments

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return compactArguments{} }
