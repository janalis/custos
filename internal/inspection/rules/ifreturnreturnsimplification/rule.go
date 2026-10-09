// Package ifreturnreturnsimplification implements the IfReturnReturnSimplification inspection.
package ifreturnreturnsimplification

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return ifReturnReturnSimplification{} }
