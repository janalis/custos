// Package ambiguousmethodscallsinarraymapping implements the AmbiguousMethodsCallsInArrayMapping inspection.
package ambiguousmethodscallsinarraymapping

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return ambiguousMethodsCallsInArrayMapping{} }
