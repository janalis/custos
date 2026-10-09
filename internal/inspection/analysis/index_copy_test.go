package analysis

import (
	"testing"

	"custos/internal/semantic/index"
)

func TestWithIndexDoesNotChangeOriginal(t *testing.T) {
	original, err := NewEngine(nil, Config{})
	if err != nil {
		t.Fatal(err)
	}
	first, second := index.New(nil), index.New(nil)
	configured := original.WithIndex(first)
	replacement := configured.WithIndex(second)
	if original.index != nil || configured.index != first || replacement.index != second || replacement == configured {
		t.Fatal("index replacement mutated a published engine")
	}
}
