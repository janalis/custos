// Package efferentobjectcoupling implements the EfferentObjectCoupling inspection.
package efferentobjectcoupling

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return efferentObjectCoupling{} }
