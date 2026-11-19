// Package traitspropertiesconflicts implements the TraitsPropertiesConflicts inspection.
package traitspropertiesconflicts

import "github.com/janalis/custos/internal/inspection/analysis"

// New constructs the stateless inspection.
func New() analysis.Rule { return traitsPropertiesConflicts{} }
