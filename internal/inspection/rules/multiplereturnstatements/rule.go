// Package multiplereturnstatements implements the MultipleReturnStatements inspection.
package multiplereturnstatements

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return multipleReturnStatements{} }
