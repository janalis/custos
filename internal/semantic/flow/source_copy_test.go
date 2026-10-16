package flow

import (
	"testing"
	"unsafe"
)

func TestSourceRepresentationSize(t *testing.T) {
	t.Logf("Source occupies %d bytes", unsafe.Sizeof(Source{}))
}

func BenchmarkSourceUnion(b *testing.B) {
	left, right := make([]Source, 16), make([]Source, 16)
	for i := range left {
		left[i] = Source{Kind: "parameter", Parameter: 2 * i}
		right[i] = Source{Kind: "parameter", Parameter: 2*i + 1}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		union(left, right)
	}
}
