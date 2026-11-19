// Package throwrawexception implements the ThrowRawException inspection.
package throwrawexception

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return throwRawException{} }
