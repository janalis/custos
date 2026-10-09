// Package strleninemptystringcheckcontext implements the StrlenInEmptyStringCheckContext inspection.
package strleninemptystringcheckcontext

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return strlenInEmptyStringCheckContext{} }
