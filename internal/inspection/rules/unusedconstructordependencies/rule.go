// Package unusedconstructordependencies implements the UnusedConstructorDependencies inspection.
package unusedconstructordependencies

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unusedConstructorDependencies{} }
