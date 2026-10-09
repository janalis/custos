// Package incrementdecrementoperationequivalent implements the IncrementDecrementOperationEquivalent inspection.
package incrementdecrementoperationequivalent

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return incrementDecrementOperationEquivalent{} }
