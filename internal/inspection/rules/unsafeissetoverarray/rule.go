// Package unsafeissetoverarray implements the UnSafeIsSetOverArray inspection.
package unsafeissetoverarray

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unSafeIsSetOverArray{} }
