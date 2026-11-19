// Package substrusedasarrayaccess implements the SubStrUsedAsArrayAccess inspection.
package substrusedasarrayaccess

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return subStrUsedAsArrayAccess{} }
