// Package accessmodifierpresented implements the AccessModifierPresented inspection.
package accessmodifierpresented

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return accessModifierPresented{} }
