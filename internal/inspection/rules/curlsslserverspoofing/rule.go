// Package curlsslserverspoofing implements the CurlSslServerSpoofing inspection.
package curlsslserverspoofing

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return curlSslServerSpoofing{} }
