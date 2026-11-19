// Package stringnormalization implements the StringNormalization inspection.
package stringnormalization

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return stringNormalization{} }
