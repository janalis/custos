// Package strstrusedasstrpos implements the StrStrUsedAsStrPos inspection.
package strstrusedasstrpos

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return strStrUsedAsStrPos{} }
