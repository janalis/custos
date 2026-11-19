// Package staticinvocationviathis implements the StaticInvocationViaThis inspection.
package staticinvocationviathis

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return staticInvocationViaThis{} }
