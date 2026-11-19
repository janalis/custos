// Package forgottendebugoutput implements the ForgottenDebugOutput inspection.
package forgottendebugoutput

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return forgottenDebugOutput{} }
