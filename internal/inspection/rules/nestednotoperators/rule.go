// Package nestednotoperators implements the NestedNotOperators inspection.
package nestednotoperators

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return nestedNotOperators{} }
