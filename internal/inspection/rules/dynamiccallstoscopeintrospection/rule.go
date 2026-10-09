// Package dynamiccallstoscopeintrospection implements the DynamicCallsToScopeIntrospection inspection.
package dynamiccallstoscopeintrospection

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return dynamicCallsToScopeIntrospection{} }
