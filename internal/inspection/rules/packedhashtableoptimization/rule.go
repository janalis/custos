// Package packedhashtableoptimization implements the PackedHashtableOptimization inspection.
package packedhashtableoptimization

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return packedHashtableOptimization{} }
