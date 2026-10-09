// Package disconnectedforeachinstruction implements the DisconnectedForeachInstruction inspection.
package disconnectedforeachinstruction

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return disconnectedForeachInstruction{} }
