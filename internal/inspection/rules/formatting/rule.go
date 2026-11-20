// Package formatting implements the native Formatting inspection. It
// reports what custos format would change in a file.
package formatting

import "github.com/janalis/custos/internal/inspection/analysis"

// New returns the Formatting inspection.
func New() analysis.Rule { return formatting{} }
