// Package unknowninspection implements the UnknownInspection inspection.
package unknowninspection

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unknownInspection{} }
