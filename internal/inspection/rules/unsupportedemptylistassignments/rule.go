// Package unsupportedemptylistassignments implements the UnsupportedEmptyListAssignments inspection.
package unsupportedemptylistassignments

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unsupportedEmptyListAssignments{} }
