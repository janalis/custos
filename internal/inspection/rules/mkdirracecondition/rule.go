// Package mkdirracecondition implements the MkdirRaceCondition inspection.
package mkdirracecondition

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return mkdirRaceCondition{} }
