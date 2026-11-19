// Package invertedifelseconstructs implements the InvertedIfElseConstructs inspection.
package invertedifelseconstructs

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return invertedIfElseConstructs{} }
