// Package implodeargumentsorder implements the ImplodeArgumentsOrder inspection.
package implodeargumentsorder

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return implodeArgumentsOrder{} }
