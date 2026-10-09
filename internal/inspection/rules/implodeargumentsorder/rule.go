// Package implodeargumentsorder implements the ImplodeArgumentsOrder inspection.
package implodeargumentsorder

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return implodeArgumentsOrder{} }
