// Package backtickoperatorusage implements the BacktickOperatorUsage inspection.
package backtickoperatorusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return backtickOperatorUsage{} }
