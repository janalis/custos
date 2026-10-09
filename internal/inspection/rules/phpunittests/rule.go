// Package phpunittests implements the PhpUnitTests inspection.
package phpunittests

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return phpUnitTests{} }
