// Package classmockingcorrectness implements the ClassMockingCorrectness inspection.
package classmockingcorrectness

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return classMockingCorrectness{} }
