// Package argumentunpackingcanbeused implements the ArgumentUnpackingCanBeUsed inspection.
package argumentunpackingcanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return argumentUnpackingCanBeUsed{} }
