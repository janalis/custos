// Package cryptographicallysecurealgorithms implements the CryptographicallySecureAlgorithms inspection.
package cryptographicallysecurealgorithms

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return cryptographicallySecureAlgorithms{} }
