// Package unsupportedstringoffsetoperations implements the UnsupportedStringOffsetOperations inspection.
package unsupportedstringoffsetoperations

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unsupportedStringOffsetOperations{} }
