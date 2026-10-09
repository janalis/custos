// Package disallowwritingintostaticproperties implements the DisallowWritingIntoStaticProperties inspection.
package disallowwritingintostaticproperties

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return disallowWritingIntoStaticProperties{} }
