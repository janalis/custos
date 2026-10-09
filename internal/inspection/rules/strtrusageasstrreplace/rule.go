// Package strtrusageasstrreplace implements the StrTrUsageAsStrReplace inspection.
package strtrusageasstrreplace

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return strTrUsageAsStrReplace{} }
