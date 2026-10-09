// Package propertyinitializationflaws implements the PropertyInitializationFlaws inspection.
package propertyinitializationflaws

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return propertyInitializationFlaws{} }
