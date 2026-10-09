// Package cascadestringreplacement implements the CascadeStringReplacement inspection.
package cascadestringreplacement

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return cascadeStringReplacement{} }
