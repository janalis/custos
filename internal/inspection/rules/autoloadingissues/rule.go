// Package autoloadingissues implements the AutoloadingIssues inspection.
package autoloadingissues

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return autoloadingIssues{} }
