// Package datetimesettimeusage implements the DateTimeSetTimeUsage inspection.
package datetimesettimeusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return dateTimeSetTimeUsage{} }
