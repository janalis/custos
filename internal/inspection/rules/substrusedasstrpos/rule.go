// Package substrusedasstrpos implements the SubStrUsedAsStrPos inspection.
package substrusedasstrpos

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return subStrUsedAsStrPos{} }
