// Package deprecatedinioptions implements the DeprecatedIniOptions inspection.
package deprecatedinioptions

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return deprecatedIniOptions{} }
