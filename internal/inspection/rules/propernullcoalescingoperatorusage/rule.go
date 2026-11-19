// Package propernullcoalescingoperatorusage implements the ProperNullCoalescingOperatorUsage inspection.
package propernullcoalescingoperatorusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return properNullCoalescingOperatorUsage{} }
