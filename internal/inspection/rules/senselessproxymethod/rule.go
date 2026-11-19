// Package senselessproxymethod implements the SenselessProxyMethod inspection.
package senselessproxymethod

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return senselessProxyMethod{} }
