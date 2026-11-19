// Package degradedswitch implements the DegradedSwitch inspection.
package degradedswitch

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return degradedSwitch{} }
