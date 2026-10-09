// Package parameterdefaultvalueisnotnull implements the ParameterDefaultValueIsNotNull inspection.
package parameterdefaultvalueisnotnull

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return parameterDefaultValueIsNotNull{} }
