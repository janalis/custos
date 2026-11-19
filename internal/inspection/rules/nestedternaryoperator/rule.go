// Package nestedternaryoperator implements the NestedTernaryOperator inspection.
package nestedternaryoperator

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return nestedTernaryOperator{} }
