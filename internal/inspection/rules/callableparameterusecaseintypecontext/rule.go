// Package callableparameterusecaseintypecontext implements the CallableParameterUseCaseInTypeContext inspection.
package callableparameterusecaseintypecontext

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return callableParameterUseCaseInTypeContext{} }
