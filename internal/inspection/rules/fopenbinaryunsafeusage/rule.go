// Package fopenbinaryunsafeusage implements the FopenBinaryUnsafeUsage inspection.
package fopenbinaryunsafeusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return fopenBinaryUnsafeUsage{} }
