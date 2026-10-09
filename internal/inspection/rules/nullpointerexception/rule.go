// Package nullpointerexception implements the NullPointerException inspection.
package nullpointerexception

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return nullPointerException{} }
