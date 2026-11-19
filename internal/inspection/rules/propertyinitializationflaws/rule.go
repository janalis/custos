// Package propertyinitializationflaws implements the PropertyInitializationFlaws inspection.
package propertyinitializationflaws

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return propertyInitializationFlaws{} }
