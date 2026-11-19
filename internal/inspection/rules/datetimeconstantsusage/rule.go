// Package datetimeconstantsusage implements the DateTimeConstantsUsage inspection.
package datetimeconstantsusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return dateTimeConstantsUsage{} }
