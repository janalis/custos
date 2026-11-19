// Package duplicatearraykeys implements the DuplicateArrayKeys inspection.
package duplicatearraykeys

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return duplicateArrayKeys{} }
