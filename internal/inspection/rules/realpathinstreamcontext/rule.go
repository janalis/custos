// Package realpathinstreamcontext implements the RealpathInStreamContext inspection.
package realpathinstreamcontext

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return realpathInStreamContext{} }
