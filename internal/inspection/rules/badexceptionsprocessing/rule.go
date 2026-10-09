// Package badexceptionsprocessing implements the BadExceptionsProcessing inspection.
package badexceptionsprocessing

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return badExceptionsProcessing{} }
