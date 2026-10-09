// Package stringsfirstcharacterscompare implements the StringsFirstCharactersCompare inspection.
package stringsfirstcharacterscompare

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return stringsFirstCharactersCompare{} }
