// Package unusedgotolabel implements the UnusedGotoLabel inspection.
package unusedgotolabel

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unusedGotoLabel{} }
