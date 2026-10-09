// Package selfclassreferencing implements the SelfClassReferencing inspection.
package selfclassreferencing

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return selfClassReferencing{} }
