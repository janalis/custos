// Package classconstantcanbeused implements the ClassConstantCanBeUsed inspection.
package classconstantcanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return classConstantCanBeUsed{} }
