// Package strtrusageasstrreplace implements the StrTrUsageAsStrReplace inspection.
package strtrusageasstrreplace

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return strTrUsageAsStrReplace{} }
