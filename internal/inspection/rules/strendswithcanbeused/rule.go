// Package strendswithcanbeused implements the StrEndsWithCanBeUsed inspection.
package strendswithcanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return strEndsWithCanBeUsed{} }
