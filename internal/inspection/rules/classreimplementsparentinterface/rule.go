// Package classreimplementsparentinterface implements the ClassReImplementsParentInterface inspection.
package classreimplementsparentinterface

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return classReImplementsParentInterface{} }
