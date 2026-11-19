// Package usinginclusiononcereturnvalue implements the UsingInclusionOnceReturnValue inspection.
package usinginclusiononcereturnvalue

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return usingInclusionOnceReturnValue{} }
