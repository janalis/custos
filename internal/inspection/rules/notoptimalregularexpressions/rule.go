// Package notoptimalregularexpressions implements the NotOptimalRegularExpressions inspection.
package notoptimalregularexpressions

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return notOptimalRegularExpressions{} }
