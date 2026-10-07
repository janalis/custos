package rules

import (
	"testing"

	"custos/internal/meta"
)

func TestRegistryConsistency(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range All() {
		m, ok := meta.Lookup(r.ID())
		if !ok || m.ID != r.ID() {
			t.Errorf("rule %q is not a catalogue ID", r.ID())
		}
		if seen[r.ID()] {
			t.Errorf("rule %q registered twice", r.ID())
		}
		seen[r.ID()] = true
	}
}
