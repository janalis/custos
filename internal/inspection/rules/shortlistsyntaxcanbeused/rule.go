// Package shortlistsyntaxcanbeused implements the ShortListSyntaxCanBeUsed inspection.
package shortlistsyntaxcanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return shortListSyntaxCanBeUsed{} }
