// Package classconstantcanbeused implements the ClassConstantCanBeUsed inspection.
package classconstantcanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return classConstantCanBeUsed{} }
