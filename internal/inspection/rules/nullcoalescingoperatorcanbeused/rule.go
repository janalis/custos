// Package nullcoalescingoperatorcanbeused implements the NullCoalescingOperatorCanBeUsed inspection.
package nullcoalescingoperatorcanbeused

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return nullCoalescingOperatorCanBeUsed{} }
