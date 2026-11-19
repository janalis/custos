// Package magicmethodsvalidity implements the MagicMethodsValidity inspection.
package magicmethodsvalidity

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return magicMethodsValidity{} }
