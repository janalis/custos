// Package uselessreturn implements the UselessReturn inspection.
package uselessreturn

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return uselessReturn{} }
