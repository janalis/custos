// Package phpunittests implements the PhpUnitTests inspection.
package phpunittests

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return phpUnitTests{} }
