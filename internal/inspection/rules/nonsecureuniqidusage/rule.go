// Package nonsecureuniqidusage implements the NonSecureUniqidUsage inspection.
package nonsecureuniqidusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return nonSecureUniqidUsage{} }
