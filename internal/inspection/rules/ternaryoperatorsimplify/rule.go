// Package ternaryoperatorsimplify implements the TernaryOperatorSimplify inspection.
package ternaryoperatorsimplify

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return ternaryOperatorSimplify{} }
