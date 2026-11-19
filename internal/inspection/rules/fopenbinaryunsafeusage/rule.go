// Package fopenbinaryunsafeusage implements the FopenBinaryUnsafeUsage inspection.
package fopenbinaryunsafeusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return fopenBinaryUnsafeUsage{} }
