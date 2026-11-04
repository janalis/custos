package infer

import (
	"testing"

	"custos/internal/semantic/types"
)

func TestNarrowingJoinKeepsUnknown(t *testing.T) {
	if !unionNarrowed(types.Unknown, types.Int).IsUnknown() {
		t.Fatal("unknown alternative supplied type evidence")
	}
	if !unionNarrowed(types.Bottom, types.Int).Equal(types.Int) {
		t.Fatal("impossible alternative poisoned the join")
	}
}
