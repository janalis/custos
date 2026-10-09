package catalogue

import (
	"testing"

	"custos/internal/inspection/meta"
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

// Catalogue order also determines which conflicting fixes get considered first.
func TestStableExecutionOrder(t *testing.T) {
	groups := []string{"Architecture", "Code style", "Compatibility", "Confusing constructs", "Control flow", "Language level migration", "Performance", "PHPUnit", "Probable bugs", "Security", "Type compatibility", "Unused"}
	ranks := make(map[string]int, len(groups))
	for rank, group := range groups {
		ranks[group] = rank
	}
	previousRank, previousID := -1, ""
	for _, rule := range All() {
		fact, _ := meta.Lookup(rule.ID())
		rank, ok := ranks[fact.Group]
		if !ok || rank < previousRank || (rank == previousRank && rule.ID() <= previousID) {
			t.Fatalf("unstable execution order: %s/%s after %s", fact.Group, rule.ID(), previousID)
		}
		previousRank, previousID = rank, rule.ID()
	}
}
