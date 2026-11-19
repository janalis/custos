// Package accessmodifierpresented implements the AccessModifierPresented inspection.
package accessmodifierpresented

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return accessModifierPresented{} }
