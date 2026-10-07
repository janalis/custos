// Package typecompat holds the rules of the "Type compatibility" group.
package typecompat

import (
	"sort"

	"custos/internal/analysis"
)

var registered []analysis.Rule

// register adds a rule to the group; call it from an init() in the rule's file.
func register(r analysis.Rule) { registered = append(registered, r) }

// Rules returns the group's rules sorted by ID.
func Rules() []analysis.Rule {
	out := append([]analysis.Rule(nil), registered...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}
