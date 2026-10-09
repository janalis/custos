// Package nonsecureuniqidusage implements the NonSecureUniqidUsage inspection.
package nonsecureuniqidusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return nonSecureUniqidUsage{} }
