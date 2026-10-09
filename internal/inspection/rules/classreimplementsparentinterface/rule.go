// Package classreimplementsparentinterface implements the ClassReImplementsParentInterface inspection.
package classreimplementsparentinterface

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return classReImplementsParentInterface{} }
