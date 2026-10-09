// Package nestedternaryoperator implements the NestedTernaryOperator inspection.
package nestedternaryoperator

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return nestedTernaryOperator{} }
