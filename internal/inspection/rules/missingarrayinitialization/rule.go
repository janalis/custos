// Package missingarrayinitialization implements the MissingArrayInitialization inspection.
package missingarrayinitialization

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return missingArrayInitialization{} }
