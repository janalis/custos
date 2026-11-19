// Package multiassignmentusage implements the MultiAssignmentUsage inspection.
package multiassignmentusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return multiAssignmentUsage{} }
