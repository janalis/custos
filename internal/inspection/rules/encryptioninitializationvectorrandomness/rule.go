// Package encryptioninitializationvectorrandomness implements the EncryptionInitializationVectorRandomness inspection.
package encryptioninitializationvectorrandomness

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return encryptionInitializationVectorRandomness{} }
