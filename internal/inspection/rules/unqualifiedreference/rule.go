// Package unqualifiedreference implements the UnqualifiedReference inspection.
package unqualifiedreference

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unqualifiedReference{} }
