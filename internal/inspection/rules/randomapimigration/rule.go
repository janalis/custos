// Package randomapimigration implements the RandomApiMigration inspection.
package randomapimigration

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return randomAPIMigration{} }
