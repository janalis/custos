// Package strendswithcanbeused implements the StrEndsWithCanBeUsed inspection.
package strendswithcanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return strEndsWithCanBeUsed{} }
