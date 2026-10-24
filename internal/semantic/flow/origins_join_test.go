package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

// Joining producers with a path whose value has an unknown producer the set
// lacks makes a new set that keeps the ambiguity.
func TestOriginsJoinKeepsOtherProducers(t *testing.T) {
	a := &syntax.Variable{Name: "a"}
	known := Value{origins: &originSet{items: []Origin{{Expr: a}}}}
	ambiguous := Value{origins: &originSet{other: true}}
	got := origins(known, ambiguous)
	if got == known.origins || got == nil || !got.other {
		t.Fatalf("join: got %+v, want a new set marked other", got)
	}
}
