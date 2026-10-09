// Package unnecessarydoublequotes implements the UnNecessaryDoubleQuotes inspection.
package unnecessarydoublequotes

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unNecessaryDoubleQuotes{} }
