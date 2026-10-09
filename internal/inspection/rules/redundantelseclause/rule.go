// Package redundantelseclause implements the RedundantElseClause inspection.
package redundantelseclause

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return redundantElseClause{} }
