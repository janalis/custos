// Package misorderedmodifiers implements the MisorderedModifiers inspection.
package misorderedmodifiers

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return misorderedModifiers{} }
