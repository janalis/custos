// Package notoptimalifconditions implements the NotOptimalIfConditions inspection.
package notoptimalifconditions

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return notOptimalIfConditions{} }
