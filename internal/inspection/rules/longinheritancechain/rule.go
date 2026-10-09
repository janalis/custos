// Package longinheritancechain implements the LongInheritanceChain inspection.
package longinheritancechain

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return longInheritanceChain{} }
