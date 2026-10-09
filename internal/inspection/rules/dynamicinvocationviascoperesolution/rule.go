// Package dynamicinvocationviascoperesolution implements the DynamicInvocationViaScopeResolution inspection.
package dynamicinvocationviascoperesolution

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return dynamicInvocationViaScopeResolution{} }
