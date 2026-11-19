// Package strstartswithcanbeused implements the StrStartsWithCanBeUsed inspection.
package strstartswithcanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return strStartsWithCanBeUsed{} }
