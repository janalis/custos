// Package incorrectrandomrange implements the IncorrectRandomRange inspection.
package incorrectrandomrange

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return incorrectRandomRange{} }
