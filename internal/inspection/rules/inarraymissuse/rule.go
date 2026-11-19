// Package inarraymissuse implements the InArrayMissUse inspection.
package inarraymissuse

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return inArrayMissUse{} }
