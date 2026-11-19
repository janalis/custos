// Package classoverridesfieldofsuperclass implements the ClassOverridesFieldOfSuperClass inspection.
package classoverridesfieldofsuperclass

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return classOverridesFieldOfSuperClass{} }
