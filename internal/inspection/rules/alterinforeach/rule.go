// Package alterinforeach implements the AlterInForeach inspection.
package alterinforeach

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return alterInForeach{} }
