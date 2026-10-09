// Package cryptographicallysecurerandomness implements the CryptographicallySecureRandomness inspection.
package cryptographicallysecurerandomness

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return cryptographicallySecureRandomness{} }
