// Package disallowwritingintostaticproperties implements the DisallowWritingIntoStaticProperties inspection.
package disallowwritingintostaticproperties

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return disallowWritingIntoStaticProperties{} }
