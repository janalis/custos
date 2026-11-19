// Package onetimeusevariables implements the OneTimeUseVariables inspection.
package onetimeusevariables

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return oneTimeUseVariables{} }
