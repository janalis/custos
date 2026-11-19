// Package arraypushmissuse implements the ArrayPushMissUse inspection.
package arraypushmissuse

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return arrayPushMissUse{} }
