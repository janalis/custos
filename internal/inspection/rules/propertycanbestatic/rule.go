// Package propertycanbestatic implements the PropertyCanBeStatic inspection.
package propertycanbestatic

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return propertyCanBeStatic{} }
