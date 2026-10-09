// Package classmethodnamematchesfieldname implements the ClassMethodNameMatchesFieldName inspection.
package classmethodnamematchesfieldname

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return classMethodNameMatchesFieldName{} }
