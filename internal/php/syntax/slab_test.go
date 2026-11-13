package syntax

import "testing"

func TestSlabGrowthPreservesNodes(t *testing.T) {
	var s slab[int]
	first := s.next()
	*first = 42
	if s.chunk != 8 {
		t.Fatal("initial slab overallocates")
	}
	for range slabChunk * 4 {
		*s.next() = 7
	}
	if *first != 42 || s.chunk != slabChunk {
		t.Fatal("growth changed a node or exceeded cap")
	}
}

func BenchmarkSlabSparse(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var s slab[Literal]
		_ = s.next()
	}
}

func BenchmarkSlabDense(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		var s slab[Literal]
		for range 1024 {
			_ = s.next()
		}
	}
}

func TestSlabChunkTransitions(t *testing.T) {
	var s slab[int]
	for _, chunk := range []int{8, 16, 32, 64, 64} {
		_ = s.next()
		if s.chunk != chunk || len(s.buf) != chunk-1 {
			t.Fatalf("chunk %d: got %d/%d", chunk, s.chunk, len(s.buf))
		}
		for range chunk - 1 {
			_ = s.next()
		}
	}
}
