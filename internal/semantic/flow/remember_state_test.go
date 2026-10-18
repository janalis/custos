package flow

import (
	"testing"

	"custos/internal/php/syntax"
)

func TestEmptyStateVisitRemainsEmptyAfterJoin(t *testing.T) {
	file := syntax.Parse("state.php", []byte("<?php probe();"), syntax.Options{})
	node := file.Stmts[0]
	result := &scopeResult{complete: true, before: map[syntax.Node]map[uint32]State{}}
	remember(result, node, frame{})
	remember(result, node, frame{states: map[uint32]State{1: {Operation: "fopen"}}})
	if states, visited := result.before[node]; !visited || len(states) != 0 {
		t.Fatal("a later path invented state absent on the first path")
	}
}

func TestRememberedStateIsIndependentOfLaterWrites(t *testing.T) {
	file := syntax.Parse("state.php", []byte("<?php probe();"), syntax.Options{})
	node := file.Stmts[0]
	result := &scopeResult{complete: true, before: map[syntax.Node]map[uint32]State{}}
	current := frame{states: map[uint32]State{1: {Operation: "fopen"}}}
	remember(result, node, current)
	current.states[1] = State{Operation: "fclose"}
	if result.before[node][1].Operation != "fopen" {
		t.Fatal("later mutation rewrote the saved predecessor")
	}
	remember(result, node, current)
	if len(result.before[node]) != 0 {
		t.Fatal("disagreeing paths retained definite state")
	}
}

func BenchmarkRememberEmptyState(b *testing.B) {
	file := syntax.Parse("state.php", []byte("<?php probe();"), syntax.Options{})
	node := file.Stmts[0]
	result := &scopeResult{complete: true, before: map[syntax.Node]map[uint32]State{}}
	b.ReportAllocs()
	for b.Loop() {
		clear(result.before)
		remember(result, node, frame{})
	}
}
