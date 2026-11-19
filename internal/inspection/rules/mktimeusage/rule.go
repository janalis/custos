// Package mktimeusage implements the MktimeUsage inspection.
package mktimeusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return mktimeUsage{} }
