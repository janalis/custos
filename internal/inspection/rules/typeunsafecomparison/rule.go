// Package typeunsafecomparison implements the TypeUnsafeComparison inspection.
package typeunsafecomparison

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return typeUnsafeComparison{} }
