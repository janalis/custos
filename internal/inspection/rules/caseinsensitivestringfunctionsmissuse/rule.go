// Package caseinsensitivestringfunctionsmissuse implements the CaseInsensitiveStringFunctionsMissUse inspection.
package caseinsensitivestringfunctionsmissuse

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return caseInsensitiveStringFunctionsMissUse{} }
