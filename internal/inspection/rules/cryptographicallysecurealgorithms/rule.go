// Package cryptographicallysecurealgorithms implements the CryptographicallySecureAlgorithms inspection.
package cryptographicallysecurealgorithms

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return cryptographicallySecureAlgorithms{} }
