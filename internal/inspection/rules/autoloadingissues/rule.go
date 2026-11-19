// Package autoloadingissues implements the AutoloadingIssues inspection.
package autoloadingissues

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return autoloadingIssues{} }
