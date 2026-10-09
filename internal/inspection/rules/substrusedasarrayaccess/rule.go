// Package substrusedasarrayaccess implements the SubStrUsedAsArrayAccess inspection.
package substrusedasarrayaccess

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return subStrUsedAsArrayAccess{} }
