// Package nestedpositiveifstatements implements the NestedPositiveIfStatements inspection.
package nestedpositiveifstatements

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return nestedPositiveIfStatements{} }
