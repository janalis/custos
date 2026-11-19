// Package referencingobjects implements the ReferencingObjects inspection.
package referencingobjects

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return referencingObjects{} }
