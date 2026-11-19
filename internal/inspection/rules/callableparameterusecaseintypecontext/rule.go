// Package callableparameterusecaseintypecontext implements the CallableParameterUseCaseInTypeContext inspection.
package callableparameterusecaseintypecontext

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return callableParameterUseCaseInTypeContext{} }
