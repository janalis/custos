// Package referencingobjects implements the ReferencingObjects inspection.
package referencingobjects

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return referencingObjects{} }
