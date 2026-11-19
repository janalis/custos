// Package shortlistsyntaxcanbeused implements the ShortListSyntaxCanBeUsed inspection.
package shortlistsyntaxcanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return shortListSyntaxCanBeUsed{} }
