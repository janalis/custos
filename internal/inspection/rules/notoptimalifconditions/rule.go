// Package notoptimalifconditions implements the NotOptimalIfConditions inspection.
package notoptimalifconditions

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return notOptimalIfConditions{} }
