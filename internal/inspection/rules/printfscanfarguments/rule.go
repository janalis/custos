// Package printfscanfarguments implements the PrintfScanfArguments inspection.
package printfscanfarguments

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return printfScanfArguments{} }
