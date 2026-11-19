// Package phpunitdeprecations implements the PhpUnitDeprecations inspection.
package phpunitdeprecations

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return phpUnitDeprecations{} }
