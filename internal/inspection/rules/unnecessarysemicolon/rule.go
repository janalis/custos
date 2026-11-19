// Package unnecessarysemicolon implements the UnnecessarySemicolon inspection.
package unnecessarysemicolon

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unnecessarySemicolon{} }
