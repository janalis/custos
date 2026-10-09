// Package misorderedmodifiers implements the MisorderedModifiers inspection.
package misorderedmodifiers

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return misorderedModifiers{} }
