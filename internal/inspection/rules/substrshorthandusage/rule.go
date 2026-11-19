// Package substrshorthandusage implements the SubStrShortHandUsage inspection.
package substrshorthandusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return subStrShortHandUsage{} }
