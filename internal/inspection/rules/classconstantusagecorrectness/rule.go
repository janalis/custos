// Package classconstantusagecorrectness implements the ClassConstantUsageCorrectness inspection.
package classconstantusagecorrectness

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return classConstantUsageCorrectness{} }
