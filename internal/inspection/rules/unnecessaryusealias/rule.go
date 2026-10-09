// Package unnecessaryusealias implements the UnnecessaryUseAlias inspection.
package unnecessaryusealias

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unnecessaryUseAlias{} }
