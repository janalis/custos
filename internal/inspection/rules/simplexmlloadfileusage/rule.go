// Package simplexmlloadfileusage implements the SimpleXmlLoadFileUsage inspection.
package simplexmlloadfileusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return simpleXMLLoadFileUsage{} }
