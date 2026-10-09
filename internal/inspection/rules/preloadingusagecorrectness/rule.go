// Package preloadingusagecorrectness implements the PreloadingUsageCorrectness inspection.
package preloadingusagecorrectness

import "custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return preloadingUsageCorrectness{} }
