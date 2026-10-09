// Package gettypemissuse implements the GetTypeMissUse inspection.
package gettypemissuse

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return getTypeMissUse{} }
