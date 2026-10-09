// Package securityadvisories implements the SecurityAdvisories inspection.
package securityadvisories

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return securityAdvisories{} }
