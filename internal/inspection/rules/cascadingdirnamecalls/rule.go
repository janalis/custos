// Package cascadingdirnamecalls implements the CascadingDirnameCalls inspection.
package cascadingdirnamecalls

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return cascadingDirnameCalls{} }
