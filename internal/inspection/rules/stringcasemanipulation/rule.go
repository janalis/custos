// Package stringcasemanipulation implements the StringCaseManipulation inspection.
package stringcasemanipulation

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return stringCaseManipulation{} }
