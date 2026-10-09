// Package invertedifelseconstructs implements the InvertedIfElseConstructs inspection.
package invertedifelseconstructs

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return invertedIfElseConstructs{} }
