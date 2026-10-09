// Package untrustedinclusion implements the UntrustedInclusion inspection.
package untrustedinclusion

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return untrustedInclusion{} }
