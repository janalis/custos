// Package senselessternaryoperator implements the SenselessTernaryOperator inspection.
package senselessternaryoperator

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return senselessTernaryOperator{} }
