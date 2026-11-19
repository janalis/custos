// Package dynamiccallstoscopeintrospection implements the DynamicCallsToScopeIntrospection inspection.
package dynamiccallstoscopeintrospection

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return dynamicCallsToScopeIntrospection{} }
