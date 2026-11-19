// Package unsafeissetoverarray implements the UnSafeIsSetOverArray inspection.
package unsafeissetoverarray

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return unSafeIsSetOverArray{} }
