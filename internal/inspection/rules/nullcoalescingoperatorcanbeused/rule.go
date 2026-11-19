// Package nullcoalescingoperatorcanbeused implements the NullCoalescingOperatorCanBeUsed inspection.
package nullcoalescingoperatorcanbeused

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return nullCoalescingOperatorCanBeUsed{} }
