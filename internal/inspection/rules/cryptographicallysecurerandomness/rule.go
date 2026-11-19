// Package cryptographicallysecurerandomness implements the CryptographicallySecureRandomness inspection.
package cryptographicallysecurerandomness

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return cryptographicallySecureRandomness{} }
