// Package unsetconstructscanbemerged implements the UnsetConstructsCanBeMerged inspection.
package unsetconstructscanbemerged

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unsetConstructsCanBeMerged{} }
