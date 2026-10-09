// Package isnullfunctionusage implements the IsNullFunctionUsage inspection.
package isnullfunctionusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return isNullFunctionUsage{} }
