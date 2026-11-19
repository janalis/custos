// Package longinheritancechain implements the LongInheritanceChain inspection.
package longinheritancechain

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return longInheritanceChain{} }
