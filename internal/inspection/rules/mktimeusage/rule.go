// Package mktimeusage implements the MktimeUsage inspection.
package mktimeusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return mktimeUsage{} }
