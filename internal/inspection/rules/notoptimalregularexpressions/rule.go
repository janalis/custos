// Package notoptimalregularexpressions implements the NotOptimalRegularExpressions inspection.
package notoptimalregularexpressions

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return notOptimalRegularExpressions{} }
