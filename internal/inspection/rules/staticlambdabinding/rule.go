// Package staticlambdabinding implements the StaticLambdaBinding inspection.
package staticlambdabinding

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return staticLambdaBinding{} }
