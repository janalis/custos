// Package nestedassignmentsusage implements the NestedAssignmentsUsage inspection.
package nestedassignmentsusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return nestedAssignmentsUsage{} }
