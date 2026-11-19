// Package jsonencodingapiusage implements the JsonEncodingApiUsage inspection.
package jsonencodingapiusage

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return jsonEncodingAPIUsage{} }
