// Package onlywritesonparameter implements the OnlyWritesOnParameter inspection.
package onlywritesonparameter

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return onlyWritesOnParameter{} }
