// Package missusingparentkeyword implements the MissUsingParentKeyword inspection.
package missusingparentkeyword

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return missUsingParentKeyword{} }
