// Package suspiciousassignments implements the SuspiciousAssignments inspection.
package suspiciousassignments

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return suspiciousAssignments{} }
