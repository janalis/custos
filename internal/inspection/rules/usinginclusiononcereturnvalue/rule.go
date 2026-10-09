// Package usinginclusiononcereturnvalue implements the UsingInclusionOnceReturnValue inspection.
package usinginclusiononcereturnvalue

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return usingInclusionOnceReturnValue{} }
