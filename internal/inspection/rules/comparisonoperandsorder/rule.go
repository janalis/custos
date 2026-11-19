// Package comparisonoperandsorder implements the ComparisonOperandsOrder inspection.
package comparisonoperandsorder

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return comparisonOperandsOrder{} }
