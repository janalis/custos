// Package pdoapiusage implements the PdoApiUsage inspection.
package pdoapiusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return pdoAPIUsage{} }
