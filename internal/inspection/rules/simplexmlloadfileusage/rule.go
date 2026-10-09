// Package simplexmlloadfileusage implements the SimpleXmlLoadFileUsage inspection.
package simplexmlloadfileusage

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return simpleXMLLoadFileUsage{} }
