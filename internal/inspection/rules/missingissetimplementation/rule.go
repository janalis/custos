// Package missingissetimplementation implements the MissingIssetImplementation inspection.
package missingissetimplementation

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return missingIssetImplementation{} }
