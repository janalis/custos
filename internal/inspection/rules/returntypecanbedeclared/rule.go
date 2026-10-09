// Package returntypecanbedeclared implements the ReturnTypeCanBeDeclared inspection.
package returntypecanbedeclared

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return returnTypeCanBeDeclared{} }
