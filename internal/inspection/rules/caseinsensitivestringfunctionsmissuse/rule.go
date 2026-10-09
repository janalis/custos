// Package caseinsensitivestringfunctionsmissuse implements the CaseInsensitiveStringFunctionsMissUse inspection.
package caseinsensitivestringfunctionsmissuse

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return caseInsensitiveStringFunctionsMissUse{} }
