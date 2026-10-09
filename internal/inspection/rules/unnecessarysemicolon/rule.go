// Package unnecessarysemicolon implements the UnnecessarySemicolon inspection.
package unnecessarysemicolon

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unnecessarySemicolon{} }
