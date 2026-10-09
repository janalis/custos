// Package typeunsafearraysearch implements the TypeUnsafeArraySearch inspection.
package typeunsafearraysearch

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return typeUnsafeArraySearch{} }
