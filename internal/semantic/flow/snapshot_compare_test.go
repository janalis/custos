package flow

import (
	"sync"
	"testing"
)

// sameHeld answers as sameSummary and compares a pair of held contracts
// once per snapshot, from any number of files at a time.
func TestSnapshotSameHeldRemembersPairs(t *testing.T) {
	s := NewSnapshotWithOptions(Options{})
	was := &Summary{Return: Value{Complete: true}, Sinks: []Sink{{Name: "echo"}}}
	equal := &Summary{Return: Value{Complete: true}, Sinks: []Sink{{Name: "echo"}}}
	other := &Summary{Return: Value{Complete: true}, Never: true}
	if !s.sameHeld(was, was) || len(s.compared.same) != 0 {
		t.Fatal("one contract compared with itself")
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			if !s.sameHeld(was, equal) || s.sameHeld(was, other) {
				t.Error("memo changed the comparison")
			}
		})
	}
	wg.Wait()
	if len(s.compared.same) != 2 || !s.compared.same[[2]*Summary{was, equal}] || s.compared.same[[2]*Summary{was, other}] {
		t.Fatalf("remembered %v", s.compared.same)
	}
	// A remembered answer is given without comparing again.
	s.compared.same[[2]*Summary{was, other}] = true
	if !s.sameHeld(was, other) {
		t.Fatal("the pair was compared again")
	}
}

func BenchmarkSnapshotSameHeld(b *testing.B) {
	s := NewSnapshotWithOptions(Options{})
	was := &Summary{Return: Value{Complete: true, Sources: []Source{{Kind: "parameter"}}}, ByRef: map[int]Value{0: {Complete: true}}}
	is := &Summary{Return: Value{Complete: true, Sources: []Source{{Kind: "parameter"}}}, ByRef: map[int]Value{0: {Complete: true}}}
	b.Run("remembered", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if !s.sameHeld(was, is) {
				b.Fatal("differ")
			}
		}
	})
	b.Run("deep", func(b *testing.B) {
		b.ReportAllocs()
		for b.Loop() {
			if !sameSummary(was, is) {
				b.Fatal("differ")
			}
		}
	})
}
