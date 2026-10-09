// Package dateintervalspecification implements the DateIntervalSpecification inspection.
package dateintervalspecification

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return dateIntervalSpecification{} }
