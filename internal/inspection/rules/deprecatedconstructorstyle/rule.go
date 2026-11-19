// Package deprecatedconstructorstyle implements the DeprecatedConstructorStyle inspection.
package deprecatedconstructorstyle

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return deprecatedConstructorStyle{} }
