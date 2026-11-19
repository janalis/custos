// Package staticlambdabinding implements the StaticLambdaBinding inspection.
package staticlambdabinding

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return staticLambdaBinding{} }
