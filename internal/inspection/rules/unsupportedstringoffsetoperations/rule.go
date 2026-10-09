// Package unsupportedstringoffsetoperations implements the UnsupportedStringOffsetOperations inspection.
package unsupportedstringoffsetoperations

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unsupportedStringOffsetOperations{} }
