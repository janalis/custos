// Package alterinforeach implements the AlterInForeach inspection.
package alterinforeach

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return alterInForeach{} }
