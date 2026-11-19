// Package stringsfirstcharacterscompare implements the StringsFirstCharactersCompare inspection.
package stringsfirstcharacterscompare

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return stringsFirstCharactersCompare{} }
