// Package untrustedinclusion implements the UntrustedInclusion inspection.
package untrustedinclusion

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return untrustedInclusion{} }
