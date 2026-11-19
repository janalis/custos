// Package issetargumentexistence implements the IssetArgumentExistence inspection.
package issetargumentexistence

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return issetArgumentExistence{} }
