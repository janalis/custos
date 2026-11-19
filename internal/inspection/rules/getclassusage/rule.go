// Package getclassusage implements the GetClassUsage inspection.
package getclassusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return getClassUsage{} }
