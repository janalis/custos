// Package nestedassignmentsusage implements the NestedAssignmentsUsage inspection.
package nestedassignmentsusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return nestedAssignmentsUsage{} }
