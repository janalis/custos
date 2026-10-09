// Package missingoremptygroupstatement implements the MissingOrEmptyGroupStatement inspection.
package missingoremptygroupstatement

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return missingOrEmptyGroupStatement{} }
