// Package opassignshortsyntax implements the OpAssignShortSyntax inspection.
package opassignshortsyntax

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return opAssignShortSyntax{} }
