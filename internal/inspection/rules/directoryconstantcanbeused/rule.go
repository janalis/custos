// Package directoryconstantcanbeused implements the DirectoryConstantCanBeUsed inspection.
package directoryconstantcanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return directoryConstantCanBeUsed{} }
