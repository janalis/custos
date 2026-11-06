package flow

import (
	"slices"
	"testing"
)

func TestGraphSuccessorStorage(t *testing.T) {
	g := graph{complete: true}
	empty := g.add(nil)
	first := g.add(nil, 2, 3)
	fanout := make([]int, 80)
	want := make([]uint16, 80)
	for i := range fanout {
		fanout[i] = i + 10
		want[i] = uint16(i + 10)
	}
	fanout[len(fanout)-1] = MaxBlocks - 1
	want[len(want)-1] = MaxBlocks - 1
	many := g.add(nil, fanout...)
	for range MaxBlocks - 3 {
		g.add(nil, 1, 2)
	}
	if got := g.add(nil, MaxBlocks-1); got != 0 || !g.full || g.complete || len(g.blocks) != MaxBlocks {
		t.Fatal("graph limit changed while adding a successor")
	}
	g.link(empty, 7)
	g.link(first, 8, 9)
	if !slices.Equal(g.successors(empty), []uint16{7}) || !slices.Equal(g.successors(first), []uint16{8, 9}) || !slices.Equal(g.successors(many), want) {
		t.Fatal("growth or relinking changed another block's ordered successors")
	}
	g.link(first, fanout...)
	g.link(many, 0, MaxBlocks-1)
	if !slices.Equal(g.successors(first), want) || !slices.Equal(g.successors(many), []uint16{0, MaxBlocks - 1}) {
		t.Fatal("fanout replacement changed a different successor list")
	}
	g.link(first)
	if len(g.successors(first)) != 0 || !slices.Equal(g.successors(empty), []uint16{7}) {
		t.Fatal("empty successor replacement changed another block")
	}
}
