// Package curlsslserverspoofing implements the CurlSslServerSpoofing inspection.
package curlsslserverspoofing

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return curlSslServerSpoofing{} }
