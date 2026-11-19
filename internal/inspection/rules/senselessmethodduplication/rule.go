// Package senselessmethodduplication implements the SenselessMethodDuplication inspection.
package senselessmethodduplication

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return senselessMethodDuplication{} }
