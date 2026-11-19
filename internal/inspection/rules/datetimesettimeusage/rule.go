// Package datetimesettimeusage implements the DateTimeSetTimeUsage inspection.
package datetimesettimeusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return dateTimeSetTimeUsage{} }
