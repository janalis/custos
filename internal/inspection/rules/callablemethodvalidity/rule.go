// Package callablemethodvalidity implements the CallableMethodValidity inspection.
package callablemethodvalidity

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return callableMethodValidity{} }
