// Package issetconstructscanbemerged implements the IssetConstructsCanBeMerged inspection.
package issetconstructscanbemerged

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return issetConstructsCanBeMerged{} }
