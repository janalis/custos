// Package pdoapiusage implements the PdoApiUsage inspection.
package pdoapiusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return pdoAPIUsage{} }
